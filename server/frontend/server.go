package frontend

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/locator"
	"github.com/memprofiler/memprofiler/server/metrics"
	"github.com/memprofiler/memprofiler/server/storage"
)

var _ schema.MemprofilerFrontendServer = (*server)(nil)

type server struct {
	httpServer *http.Server
	computer   metrics.Computer
	storage    storage.Storage
	errChan    chan<- error
	logger     logrus.FieldLogger
}

func (s *server) GetServices(
	ctx context.Context,
	request *schema.GetServicesRequest,
) (*schema.GetServicesResponse, error) {
	return &schema.GetServicesResponse{ServiceTypes: s.storage.Services()}, nil
}

func (s *server) GetInstances(
	ctx context.Context,
	request *schema.GetInstancesRequest,
) (*schema.GetInstancesResponse, error) {
	instances, err := s.storage.Instances(request.GetServiceType())
	if err != nil {
		// TODO: think about google.golang.org/grpc/status
		return nil, err
	}
	return &schema.GetInstancesResponse{ServiceInstances: instances}, nil
}

func (s *server) GetSessions(
	ctx context.Context,
	request *schema.GetSessionsRequest,
) (*schema.GetSessionsResponse, error) {
	sessions, err := s.storage.Sessions(request.GetServiceDescription())
	if err != nil {
		// TODO: think about google.golang.org/grpc/status
		return nil, err
	}
	return &schema.GetSessionsResponse{Sessions: sessions}, nil
}

func (s *server) SubscribeForSession(
	request *schema.SubscribeForSessionRequest,
	stream schema.MemprofilerFrontend_SubscribeForSessionServer) error {

	// make subscription for a requested service
	subscription, err := s.computer.SessionSubscribe(stream.Context(), request.GetSessionDescription())
	if err != nil {
		return err
	}

	// push session metrics to the client
	for {
		select {
		case msg, ok := <-subscription.Updates():
			if !ok {
				// session terminated by a service
				return nil
			}
			// sort trend values by InUseBytes rate, since it the most relevant indicator for memory leak
			sort.Slice(msg.Locations, func(i, j int) bool {
				// descending order
				return msg.Locations[i].Rates[0].Values.InUseBytes > msg.Locations[j].Rates[0].Values.InUseBytes
			})
			if err := stream.Send(msg); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

// Start runs HTTP Backend
func (s *server) Start() { s.errChan <- s.httpServer.ListenAndServe() }

const terminationTimeout = time.Second

// Stop terminates HTTP Backend
func (s *server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), terminationTimeout)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil && err != ctx.Err() {
		s.logger.WithError(err).Error("Backend shutdown error")
	}
}

// NewServer initializes new server
func NewServer(
	cfg *config.FrontendConfig,
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
