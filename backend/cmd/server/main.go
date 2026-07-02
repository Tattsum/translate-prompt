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
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	appintake "github.com/Tattsum/translate-prompt/backend/application/intake"
	"github.com/Tattsum/translate-prompt/backend/application/optimize"
	"github.com/Tattsum/translate-prompt/backend/graph"
	infraBP "github.com/Tattsum/translate-prompt/backend/infrastructure/bestpractice"
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

	loader, err := infraBP.NewLoader()
	if err != nil {
		log.Fatalf("load rules: %v", err)
	}

	optUC, err := optimize.NewUseCase(loader, "cl100k_base")
	if err != nil {
		log.Fatalf("optimizer: %v", err)
	}
	intakeUC := appintake.NewUseCase(loader)

	mux := http.NewServeMux()

	gqlResolver := graph.NewResolver(optUC, intakeUC)
	gqlServer := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: gqlResolver}))
	mux.Handle(graphQLEndpoint, gqlServer)

	if env := os.Getenv("ENV"); env == "development" || env == "dev" {
		mux.Handle(playgroundEndpoint, playground.Handler("translate-prompt", graphQLEndpoint))
	}

	connectrpc.Mount(mux, connectrpc.NewService(optUC, intakeUC))
	registerSPA(mux)

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	log.Printf("listening on http://%s (GraphQL %s, Connect RPC)", addr, graphQLEndpoint)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	srv := &http.Server{
		Addr:              addr,
		Handler:           corsMiddleware(mux),
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
