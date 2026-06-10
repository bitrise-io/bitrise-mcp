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
	"github.com/bitrise-io/bitrise-mcp/v2/internal/tool"
	"github.com/jinzhu/configor"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const development = "development"

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
	// EnabledAPIGroups is a comma-separated list of API groups that are enabled.
	EnabledAPIGroups string `env:"ENABLED_API_GROUPS" default:"apps,builds,workspaces,outgoing-webhooks,artifacts,group-roles,cache-items,pipelines,account,read-only,release-management,registration"`
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
	// BitriseAPIBaseURL overrides the Bitrise v0.1 API base URL
	// (default: https://api.bitrise.io/v0.1). Useful for pointing at a
	// test or local API instance.
	BitriseAPIBaseURL string `env:"BITRISE_API_BASE_URL"`
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
		server.WithToolFilter(func(ctx context.Context, tools []mcp.Tool) []mcp.Tool {
			enabledGroups, err := bitrise.EnabledGroupsFromCtx(ctx) // http transport only
			if err != nil {
				// stdio transport/no tool filtering in http transport
				enabledGroups = strings.Split(cfg.EnabledAPIGroups, ",")
			}
			var filtered []mcp.Tool
			for _, tool := range tools {
				if toolBelt.ToolEnabled(tool.Name, enabledGroups) {
					filtered = append(filtered, tool)
				}
			}
			return filtered
		}),
		server.WithRecovery(),
		server.WithToolCapabilities(false),
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
		return runStdioTransport(cfg, mcpServer)
	}
	logger.Info("starting http transport")
	return runHTTPTransport(mcpServer, logger, cfg)
}

func runStdioTransport(cfg config, mcpServer *server.MCPServer) error {
	server.WithToolHandlerMiddleware(func(fn server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return fn(bitrise.ContextWithPAT(ctx, cfg.BitriseToken), request)
		}
	})(mcpServer)
	if err := server.ServeStdio(mcpServer); err != nil {
		return fmt.Errorf("serve stdio: %w", err)
	}
	return nil
}

func runHTTPTransport(mcpServer *server.MCPServer, logger *zap.SugaredLogger, cfg config) error {
	if cfg.BitriseToken != "" {
		return fmt.Errorf("BITRISE_TOKEN cannot be provided in http transport mode")
	}

	var exchanger *jwtExchanger
	if cfg.OIDCTokenEndpoint != "" {
		exchanger = &jwtExchanger{tokenEndpoint: cfg.OIDCTokenEndpoint, logger: logger}
	}

	mcpHandler := server.NewStreamableHTTPServer(
		mcpServer,
		server.WithStateLess(true),
		server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token != "" {
				pat := token
				if exchanger != nil && isJWT(token) {
					var err error
					pat, err = exchanger.exchange(r.Context(), token)
					if err != nil {
						logger.Warnw("JWT→PAT exchange failed", "error", err)
						return ctx
					}
				}
				ctx = bitrise.ContextWithPAT(ctx, pat)
			}
			// server.WithToolFilter can use it to limit the tools listed.
			enabledGroups := r.Header.Get("x-bitrise-enabled-api-groups")
			if enabledGroups != "" {
				a := strings.Split(enabledGroups, ",")
				ctx = bitrise.ContextWithEnabledGroups(ctx, a)
			}
			return ctx
		}),
		server.WithLogger(logger),
		server.WithDisableStreaming(true),
	)

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
	if cfg.ExternalOAuthIssuer != "" {
		mux.HandleFunc("/.well-known/oauth-protected-resource", oauthProtectedResourceHandler(cfg.ExternalOAuthIssuer))
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If the request looks like it's from a browser (Sec-Fetch-Mode: navigate),
		// redirect to the documentation instead of handling as MCP request.
		if r.Header.Get("Sec-Fetch-Mode") == "navigate" {
			http.Redirect(w, r, "https://github.com/bitrise-io/bitrise-mcp/blob/main/README.md", http.StatusTemporaryRedirect)
			return
		}
		// Otherwise, handle as MCP request
		mcpHandler.ServeHTTP(w, r)
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
