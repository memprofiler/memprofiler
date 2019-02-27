package frontend

import (
	"context"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"time"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/locator"
	"github.com/memprofiler/memprofiler/server/metrics"
	"github.com/memprofiler/memprofiler/server/storage"
	"github.com/sirupsen/logrus"
)

var _ schema.MemprofilerFrontendServer = (*server)(nil)

type server struct {
	httpServer *http.Server
	computer   metrics.Computer
	storage    storage.Storage
	errChan    chan<- error
	logger     logrus.FieldLogger
}

func (s *server) GetServices(ctx context.Context, request *schema.GetServicesRequest) (*schema.GetServicesResponse, error) {
	return &schema.GetServicesResponse{ServiceKinds: s.storage.Services()}, nil
}

func (s *server) GetInstances(ctx context.Context, request *schema.GetInstancesRequest) (*schema.GetInstancesResponse, error) {
	return &schema.GetInstancesResponse{
		ServiceInstances: s.storage.Instances(request.GetServiceKind()),
	}, nil
}

func (s *server) GetSessions(ctx context.Context, request *schema.GetSessionsRequest) (*schema.GetSessionsResponse, error) {
	panic("implement me")
}

func (s *server) SubscribeForSession(ctx *schema.SubscribeForSessionRequest, request schema.MemprofilerFrontend_SubscribeForSessionServer) error {
	panic("implement me")
}

// Start runs HTTP API
func (s *server) Start() { s.errChan <- s.httpServer.ListenAndServe() }

const terminationTimeout = time.Second

// Stop terminates HTTP API
func (s *server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), terminationTimeout)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil && err != ctx.Err() {
		s.logger.WithError(err).Error("API shutdown error")
	}
}

// NewServer initializes new server
func NewServer(
	cfg *config.WebConfig,
	locator *locator.Locator,
	errChan chan<- error,
) (common.Service, error) {

	s := &server{
		computer: locator.Computer,
		storage:  locator.Storage,
		logger:   locator.Logger.WithField("subsystem", "frontend"),
		errChan:  errChan,
	}

	grpcServer := grpc.NewServer()
	schema.RegisterMemprofilerFrontendServer(grpcServer, s)
	grpclog.SetLogger(locator.Logger) // FIXME: replace to V2
	wrappedServer := grpcweb.WrapServer(grpcServer)

	// Dump to logs resource list
	for _, resource := range grpcweb.ListGRPCResources(grpcServer) {
		s.logger.WithField("URL", resource).Info("HTTP Frontend server resource")
	}

	handler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedServer.ServeHTTP(resp, req)
	}
	s.httpServer = &http.Server{
		Addr:    cfg.ListenEndpoint,
		Handler: http.HandlerFunc(handler),
	}

	return s, nil
}
