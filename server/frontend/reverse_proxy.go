package frontend

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	gw "github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/locator"
)

var _ common.Service = (*reverseProxyServer)(nil)

type reverseProxyServer struct {
	httpServer *http.Server
	errChan    chan<- error
	logger     *zerolog.Logger
	cfg        *config.FrontendConfig
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func (r *reverseProxyServer) Start() {
	r.errChan <- r.httpServer.ListenAndServe()
}

func (r *reverseProxyServer) Stop() {
	defer r.cancelFunc()

	if err := r.httpServer.Shutdown(r.ctx); err != nil && err != r.ctx.Err() {
		r.logger.Err(err)
	}
}

// startReverseProxy start reverse proxy for front
func NewReverseProxy(
	cfg *config.FrontendConfig,
	locator *locator.Locator,
	errChan chan<- error,
) (common.Service, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// register gRPC server endpoint
	var (
		opts       = []grpc.DialOption{grpc.WithInsecure()}
		mux        = http.NewServeMux()
		runtimeMux = runtime.NewServeMux()
		fs         = http.FileServer(http.Dir("www/swagger-ui"))
	)

	err := gw.RegisterMemprofilerFrontendHandlerFromEndpoint(ctx, runtimeMux, cfg.ListenEndpoint, opts)
	if err != nil {
		defer cancel()

		return nil, err
	}

	server := &http.Server{Addr: cfg.FrontendEndpoint, Handler: allowCORS(mux)}

	// Serve the swagger-ui-ui and swagger-ui file
	mux.Handle("/", runtimeMux)
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui", fs))
	mux.HandleFunc("/swagger.json", serveSwagger)

	// start HTTP server (and proxy calls to gRPC server endpoint)
	return &reverseProxyServer{
		httpServer: server,
		errChan:    errChan,
		logger:     locator.Logger,
		cfg:        cfg,
		ctx:        ctx,
		cancelFunc: cancel,
	}, nil
}

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "www/swagger.json")
}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))

	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))

	glog.Infof("preflight request for %s", r.URL.Path)

	return
}

// allowCORS allows Cross Origin Resoruce Sharing from any origin.
// Don't do this without consideration in production systems.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)

			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)

				return
			}
		}

		h.ServeHTTP(w, r)
	})
}
