package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	appintake "github.com/Tattsum/translate-prompt/backend/application/intake"
	"github.com/Tattsum/translate-prompt/backend/application/optimize"
	"github.com/Tattsum/translate-prompt/backend/domain/budget"
	"github.com/Tattsum/translate-prompt/backend/graph"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
	"github.com/Tattsum/translate-prompt/backend/infrastructure/config"
	infrallm "github.com/Tattsum/translate-prompt/backend/infrastructure/llm"
	"github.com/Tattsum/translate-prompt/backend/presentation/connectrpc"
)

const (
	graphQLEndpoint    = "/query"
	playgroundEndpoint = "/playground"
	readHeaderTimeout  = 10 * time.Second
)

func main() {
	port := flag.Int("port", 8080, "listen port")
	flag.Parse()

	cfg := config.LoadServerFromEnv(*port)
	llmCfg := infrallm.LoadConfigFromEnv()
	llmCfg.Enabled = llmCfg.Enabled || cfg.LLM.Enabled

	loader, err := infraBP.NewLoader()
	if err != nil {
		log.Fatalf("load rules: %v", err)
	}

	llmService := infrallm.NewService(llmCfg)

	optUC, err := optimize.NewUseCase(loader, "cl100k_base")
	if err != nil {
		log.Fatalf("optimizer: %v", err)
	}
	optUC.WithCompleter(llmService)
	intakeUC := appintake.NewUseCase(loader).WithCompleter(llmService)

	mux := http.NewServeMux()

	gqlResolver := graph.NewResolver(optUC, intakeUC, cfg.InvestigateEnabled, llmBudgetDefaults(llmCfg))
	gqlServer := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: gqlResolver}))
	mux.Handle(graphQLEndpoint, gqlServer)

	if env := os.Getenv("ENV"); env == "development" || env == "dev" {
		mux.Handle(playgroundEndpoint, playground.Handler("translate-prompt", graphQLEndpoint))
	}

	connectrpc.Mount(mux, connectrpc.NewService(optUC, intakeUC, cfg.InvestigateEnabled, llmBudgetDefaults(llmCfg)))
	registerSPA(mux)

	addr := fmt.Sprintf("%s:%d", cfg.ListenHost, cfg.Port)
	log.Printf("listening on http://%s (GraphQL %s, Connect RPC)", addr, graphQLEndpoint)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	srv := &http.Server{
		Addr:              addr,
		Handler:           corsMiddleware(cfg.AllowedOrigins, mux),
		ReadHeaderTimeout: readHeaderTimeout,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func llmBudgetDefaults(cfg infrallm.Config) budget.Config {
	out := budget.DefaultConfig()
	cfg.ApplyToBudgetConfig(&out)
	return out
}

func corsMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	allowAll := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowAll {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")

		if r.Method == http.MethodOptions {
			if !allowAll && origin != "" {
				if _, ok := allowed[origin]; !ok {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		return true
	}
	for _, allowed := range allowedOrigins {
		if strings.EqualFold(allowed, origin) {
			return true
		}
	}
	return false
}
