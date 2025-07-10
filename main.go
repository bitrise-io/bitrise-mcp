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

	"github.com/bitrise-io/bitrise-mcp/internal/tool"
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
	EnabledAPIGroups string `env:"ENABLED_API_GROUPS" default:"apps,builds,workspaces,webhooks,build-artifacts,group-roles,cache-items,pipelines,account,read-only,release-management"`
	// LogLevel is the log level for the application.
	LogLevel string `env:"LOG_LEVEL" default:"info"`
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

	logger, err := newStructuredLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}

	toolBelt := tool.NewBelt()
	mcpServer := server.NewMCPServer(
		"bitrise",
		"2.0.0",
		server.WithToolFilter(func(ctx context.Context, tools []mcp.Tool) []mcp.Tool {
			enabledGroups, err := tool.EnabledGroupsFromCtx(ctx)
			if err != nil {
				logger.Errorf("get enabled api groups from context: %v", err)
				return tools // return all tools if there's an error
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

	if cfg.Addr == "" {
		logger.Info("no address specified, starting stdio transport")
		return runStdioTransport(cfg, mcpServer)
	}
	logger.Info("starting http transport")
	return runHTTPTransport(mcpServer, logger, cfg)
}

func runStdioTransport(cfg config, mcpServer *server.MCPServer) error {
	if cfg.BitriseToken == "" {
		return fmt.Errorf("BITRISE_TOKEN must be provided in stdio transport mode")
	}

	server.WithToolHandlerMiddleware(func(fn server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ctx = tool.ContextWithPAT(ctx, cfg.BitriseToken)

			enabledGroups := strings.Split(cfg.EnabledAPIGroups, ",")
			ctx = tool.ContextWithEnabledGroups(ctx, enabledGroups)

			return fn(ctx, request)
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

	httpServer := server.NewStreamableHTTPServer(
		mcpServer,
		server.WithEndpointPath("/"),
		server.WithStateLess(true),
		server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			pat := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if pat != "" {
				ctx = tool.ContextWithPAT(ctx, pat)
			}

			enabledGroups := strings.Split(cfg.EnabledAPIGroups, ",")
			clientEnabledGroups := r.Header.Get("x-bitrise-enabled-api-groups")
			if clientEnabledGroups != "" {
				enabledGroups = strings.Split(clientEnabledGroups, ",")
			}
			ctx = tool.ContextWithEnabledGroups(ctx, enabledGroups)

			return ctx
		}),
		server.WithLogger(logger),
	)

	// Start the HTTP server in another goroutine.
	errListen := make(chan error, 1)
	go func() {
		err := httpServer.Start(cfg.Addr)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errListen <- fmt.Errorf("listen and serve: %w", err)
			return
		}
		errListen <- nil
	}()
	logger.Infof("started listening on %q\n", cfg.Addr)

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

	loggerConfig.OutputPaths = []string{"stdout"}
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerConfig.Level = atom

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("new zap logger: %w", err)
	}
	return logger.Sugar(), nil
}
