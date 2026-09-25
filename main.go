package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	httptrace "github.com/DataDog/dd-trace-go/contrib/net/http/v2"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/bitrise"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/devenv"
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool"
	"github.com/jinzhu/configor"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const development = "development"

// serverInstructions is sent to clients at initialize time: it tells the
// model how the two product areas of the server relate, since tool
// descriptions alone cannot say which product a request is about.
const serverInstructions = `Bitrise MCP server. Two product areas:
1. Bitrise CI, Release Management and Insights: unprefixed tools (list_apps, trigger_bitrise_build, get_build_log, list_connected_apps, insights_*). Apps are addressed by app_slug, workspaces by workspace_slug (organization_slug on register_app); results use snake_case fields.
2. Bitrise Dev Environments (RDE): tools prefixed bitrise_devenv_*. Remote development sessions (VMs) created from templates or a stack, with shell execution, file transfer, virtual devices and macOS GUI automation. Workspace-scoped tools take workspace_id (the workspace slug); results use camelCase fields.
trigger_bitrise_build runs a CI workflow on Bitrise; bitrise_devenv_execute runs a shell command inside an existing Dev Environments session. One Bitrise account covers both areas: me identifies the user, list_workspaces lists the workspaces (slugs) both areas use.
Dev Environments workspace resolution: an explicit workspace_id argument (remembered for later calls), else the BITRISE_WORKSPACE_ID environment variable or the x-bitrise-workspace-id header, else the sole workspace the user belongs to; with several workspaces and none chosen, ask the user. A Workspace API Token has no user and no workspace discovery: it needs the workspace given explicitly.`

// BuildVersion is overwritten with go build flags.
var BuildVersion = development //nolint:gochecknoglobals

type config struct {
	// Addr is the address to listen on for HTTP transport in host:port format.
	// If set, the server will use HTTP transport, otherwise it will use stdio
	// transport.
	Addr string `env:"ADDR"`
	// BitriseToken is the Bitrise API token used to authenticate requests for
	// the stdio transport. Only valid for the stdio transport, otherwise it is
	// ignored.
	BitriseToken string `env:"BITRISE_TOKEN"`
	// BitriseWorkspaceID is the default workspace ID (slug) the Dev
	// Environments tools operate in on the stdio transport. On the HTTP
	// transport the workspace comes from the x-bitrise-workspace-id header.
	// Optional: a workspace_id tool argument wins, and when the user belongs
	// to exactly one workspace it is auto-detected.
	BitriseWorkspaceID string `env:"BITRISE_WORKSPACE_ID"`
	// EnabledAPIGroups is a comma-separated list of API groups that are enabled.
	EnabledAPIGroups string `env:"ENABLED_API_GROUPS" default:"apps,builds,workspaces,outgoing-webhooks,artifacts,group-roles,cache-items,pipelines,account,user,read-only,release-management,release-management-code-push,configuration,insights,dev-environments,dev-environments-read-only"`
	// LogLevel is the log level for the application.
	LogLevel string `env:"LOG_LEVEL" default:"info"`
	// DatadogTracingEnabled enables DataDog APM tracing when set to true.
	// Requires a DataDog agent to be running and reachable (DD_AGENT_HOST).
	DatadogTracingEnabled bool `env:"DATADOG_TRACING_ENABLED" default:"false"`
	// ExternalOAuthIssuer is the issuer URL of an external OAuth authorization
	// server. When set, the server advertises
	// /.well-known/oauth-protected-resource so OAuth clients can discover the
	// correct authorization server. Requires OIDCTokenEndpoint to be set as well.
	ExternalOAuthIssuer string `env:"EXTERNAL_OAUTH_ISSUER"`
	// OIDCTokenEndpoint is the full URL of the OIDC token exchange endpoint
	// (RFC 8693) used to trade an external JWT for a Bitrise PAT. When set,
	// Bearer tokens that look like JWTs are exchanged before being passed to tools.
	OIDCTokenEndpoint string `env:"OIDC_TOKEN_ENDPOINT"`
	// ServerBaseURL is the public base URL of this MCP server (e.g. https://mcp.bitrise.io).
	// Required when EXTERNAL_OAUTH_ISSUER is set — used in WWW-Authenticate headers and
	// the OAuth protected resource metadata document.
	ServerBaseURL string `env:"SERVER_BASE_URL"`
	// BitriseAPIBaseURL overrides the Bitrise v0.1 API base URL
	// (default: https://api.bitrise.io/v0.1). Useful for pointing at a
	// test or local API instance.
	BitriseAPIBaseURL string `env:"BITRISE_API_BASE_URL"`
	// DevenvAPIBaseURL is the base URL of the Dev Environments (RDE) backend
	// used by the bitrise_devenv_* tools.
	DevenvAPIBaseURL string `env:"BITRISE_DEVENV_API_BASE_URL" default:"https://codespaces-api.services.bitrise.io"`
}

// splitGroups parses a comma-separated API group list, ignoring surrounding
// whitespace and empty items.
func splitGroups(s string) []string {
	var groups []string
	for _, g := range strings.Split(s, ",") {
		if g = strings.TrimSpace(g); g != "" {
			groups = append(groups, g)
		}
	}
	return groups
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %+v", err)
	}
}

func run() error {
	var cfg config
	if err := configor.Load(&cfg); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	if cfg.BitriseAPIBaseURL != "" {
		bitrise.APIBaseURL = cfg.BitriseAPIBaseURL
	}
	devenv.BaseURL = strings.TrimRight(cfg.DevenvAPIBaseURL, "/")

	logger, err := newStructuredLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}

	if cfg.DatadogTracingEnabled {
		err := tracer.Start(
			tracer.WithService("bitrise-mcp"),
			tracer.WithServiceVersion(BuildVersion),
		)
		if err != nil {
			log.Fatalf("Unable to start tracing: %s", err)
		}
		defer tracer.Stop()
	}

	toolBelt := tool.NewBelt()
	mcpServer := server.NewMCPServer(
		"bitrise",
		BuildVersion,
		server.WithInstructions(serverInstructions),
		server.WithToolFilter(toolBelt.ToolFilter(splitGroups(cfg.EnabledAPIGroups))),
		server.WithRecovery(),
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
		server.WithLogging(),
	)
	toolBelt.RegisterAll(mcpServer)

	if cfg.DatadogTracingEnabled {
		transport := "http"
		if cfg.Addr == "" {
			transport = "stdio"
		}
		server.WithToolHandlerMiddleware(func(fn server.ToolHandlerFunc) server.ToolHandlerFunc {
			return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				span, ctx := tracer.StartSpanFromContext(ctx, "mcp.tool",
					tracer.ResourceName(request.Params.Name),
					tracer.SpanType("rpc"),
					tracer.Tag("mcp.tool", request.Params.Name),
					tracer.Tag("mcp.transport", transport),
				)

				result, err := fn(ctx, request)

				if err != nil {
					span.Finish(tracer.WithError(err))
					return result, err
				}
				// Call itself was successful but the result is an error
				if result != nil && result.IsError {
					span.SetTag("mcp.tool.is_error", true)
				}
				span.Finish()

				return result, nil
			}
		})(mcpServer)
	}

	if cfg.Addr == "" {
		logger.Info("no address specified, starting stdio transport")
		return runStdioTransport(cfg, toolBelt, mcpServer)
	}
	logger.Info("starting http transport")
	return runHTTPTransport(mcpServer, toolBelt, logger, cfg)
}

func runStdioTransport(cfg config, toolBelt *tool.Belt, mcpServer *server.MCPServer) error {
	if cfg.BitriseToken == "" {
		return fmt.Errorf("BITRISE_TOKEN must be provided in stdio transport mode")
	}

	server.WithToolHandlerMiddleware(func(fn server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ctx = bitrise.ContextWithPAT(ctx, cfg.BitriseToken)
			if cfg.BitriseWorkspaceID != "" {
				ctx = devenv.ContextWithWorkspace(ctx, cfg.BitriseWorkspaceID)
			}
			return fn(ctx, request)
		}
	})(mcpServer)
	server.WithToolHandlerMiddleware(workspaceGateMiddleware(toolBelt))(mcpServer)
	if err := server.ServeStdio(mcpServer); err != nil {
		return fmt.Errorf("serve stdio: %w", err)
	}
	return nil
}

// workspaceGateMiddleware runs the Dev Environments per-call gate: it rejects
// local-only tools on the hosted transport and resolves the workspace for
// workspace-scoped bitrise_devenv_* tools. Every other tool passes through.
// It must be registered after the middleware that puts the PAT in context.
func workspaceGateMiddleware(toolBelt *tool.Belt) server.ToolHandlerMiddleware {
	return func(fn server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ctx, errRes := toolBelt.GateAndResolveWorkspace(ctx, request)
			if errRes != nil {
				return errRes, nil
			}
			return fn(ctx, request)
		}
	}
}

func runHTTPTransport(mcpServer *server.MCPServer, toolBelt *tool.Belt, logger *zap.SugaredLogger, cfg config) error {
	if cfg.BitriseToken != "" {
		return fmt.Errorf("BITRISE_TOKEN cannot be provided in http transport mode")
	}
	if cfg.ExternalOAuthIssuer != "" && (cfg.OIDCTokenEndpoint == "" || cfg.ServerBaseURL == "") {
		return fmt.Errorf("EXTERNAL_OAUTH_ISSUER requires OIDC_TOKEN_ENDPOINT and SERVER_BASE_URL to be set")
	}

	var exchanger *jwtExchanger
	if cfg.OIDCTokenEndpoint != "" {
		exchanger = &jwtExchanger{tokenEndpoint: cfg.OIDCTokenEndpoint}
	}

	// When an external OAuth issuer is configured the auth middleware challenges
	// credential-less requests at the connection layer (returning a 401 +
	// WWW-Authenticate). The WithHTTPContextFunc still resolves the bearer token
	// to a PAT for authenticated requests. Otherwise (no issuer) the context func
	// is the only auth step, preserving the previous behaviour where missing auth
	// surfaces as an in-band tool error.
	externalOAuthConfigured := cfg.ExternalOAuthIssuer != ""

	var metadataURL string
	httpServerOpts := []server.StreamableHTTPOption{
		server.WithStateLess(true),
		server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			ctx = devenv.ContextWithHostedMode(ctx)
			pat, err := extractPAT(r, exchanger)
			if err != nil {
				logger.Warnw("JWT→PAT exchange failed", "error", err)
			} else if pat != "" {
				ctx = bitrise.ContextWithPAT(ctx, pat)
			}
			// server.WithToolFilter can use it to limit the tools listed.
			if enabledGroups := r.Header.Get("x-bitrise-enabled-api-groups"); enabledGroups != "" {
				ctx = bitrise.ContextWithEnabledGroups(ctx, splitGroups(enabledGroups))
			}
			// Default workspace for the Dev Environments tools.
			if ws := r.Header.Get("x-bitrise-workspace-id"); ws != "" {
				ctx = devenv.ContextWithWorkspace(ctx, ws)
			}
			return ctx
		}),
		server.WithDisableStreaming(true),
	}
	if externalOAuthConfigured {
		protectedResourceCfg := server.ProtectedResourceMetadataConfig{
			Resource:               cfg.ServerBaseURL,
			AuthorizationServers:   []string{cfg.ExternalOAuthIssuer},
			BearerMethodsSupported: []string{"header"},
		}
		httpServerOpts = append(httpServerOpts, server.WithProtectedResourceMetadata(protectedResourceCfg))
		metadataURL = cfg.ServerBaseURL + server.WellKnownProtectedResourcePath
	}
	// PAT and the header workspace are already in context from the HTTP
	// context func above; gate the Dev Environments tools per call.
	server.WithToolHandlerMiddleware(workspaceGateMiddleware(toolBelt))(mcpServer)

	mcpHandler := server.NewStreamableHTTPServer(mcpServer, httpServerOpts...)

	type router interface {
		http.Handler
		HandleFunc(string, func(http.ResponseWriter, *http.Request))
	}

	var mux router
	if cfg.DatadogTracingEnabled {
		mux = httptrace.NewServeMux()
	} else {
		mux = http.NewServeMux()
	}
	mux.HandleFunc("/readyz", readyzHandler)
	mux.HandleFunc("/livez", livezHandler)

	var mcpEntry http.Handler = mcpHandler
	if externalOAuthConfigured {
		mcpEntry = requireAuthMiddleware(mcpHandler, exchanger, metadataURL, logger)
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If the request looks like it's from a browser (Sec-Fetch-Mode: navigate),
		// redirect to the documentation instead of handling as MCP request.
		if r.Header.Get("Sec-Fetch-Mode") == "navigate" {
			http.Redirect(w, r, "https://github.com/bitrise-io/bitrise-mcp/blob/main/README.md", http.StatusTemporaryRedirect)
			return
		}
		// Otherwise, handle as MCP request
		mcpEntry.ServeHTTP(w, r)
	})

	httpServer := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	// Start the HTTP server in another goroutine.
	errListen := make(chan error, 1)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errListen <- fmt.Errorf("listen and serve: %w", err)
			return
		}
		errListen <- nil
	}()
	logger.Infof("started listening on %q", cfg.Addr)

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// In main goroutine, wait for either...
	select {
	case <-ctx.Done():
		// ... signal for operating system to terminate.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		// Terminate net/http server with a grace period.
		logger.Infoln("shutting down http server")
		if err := httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		logger.Infoln("http server shutdown successful")
	case err := <-errListen:
		// ... error of net/http server.
		return err
	}
	return nil
}

func newStructuredLogger(level string) (*zap.SugaredLogger, error) {
	atom := zap.NewAtomicLevel()
	if err := atom.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("could parse log level: %w", err)
	}

	loggerConfig := zap.NewProductionConfig()
	if BuildVersion == development {
		loggerConfig = zap.NewDevelopmentConfig()
		loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		loggerConfig.DisableStacktrace = true
	}

	loggerConfig.OutputPaths = []string{"stderr"}
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerConfig.Level = atom

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("new zap logger: %w", err)
	}
	return logger.Sugar(), nil
}

func readyzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func livezHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
