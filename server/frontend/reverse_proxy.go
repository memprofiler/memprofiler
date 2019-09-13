package frontend

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	gw "github.com/memprofiler/memprofiler/schema"
)

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "www/swagger.json")
}

// StartReverseProxy start reverse proxy for front
func StartReverseProxy(backendEndpoint, frontendEndpoint string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// register gRPC server endpoint
	rmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterMemprofilerFrontendHandlerFromEndpoint(ctx, rmux, backendEndpoint, opts)
	if err != nil {
		return err
	}

	// Serve the swagger-ui-ui and swagger-ui file
	mux := http.NewServeMux()
	mux.Handle("/", rmux)
	mux.HandleFunc("/swagger.json", serveSwagger)
	fs := http.FileServer(http.Dir("www/swagger-ui"))
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui", fs))

	// start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(frontendEndpoint, mux)
}
