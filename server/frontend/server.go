package frontend

import (
	"context"
	"net"
	"sort"

	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/common"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/locator"
	"github.com/memprofiler/memprofiler/server/metrics"
	"github.com/memprofiler/memprofiler/server/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var _ schema.MemprofilerFrontendServer = (*server)(nil)

type server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	computer   metrics.Computer
	storage    storage.Storage
	errChan    chan<- error
	logger     *zerolog.Logger
	cfg        *config.FrontendConfig
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

func (s *server) Start() {
	s.errChan <- s.grpcServer.Serve(s.listener)
}

func (s *server) Stop() { s.grpcServer.GracefulStop() }

// NewServer builds new GRPC server
func NewServer(
	cfg *config.FrontendConfig,
	locator *locator.Locator,
	errChan chan<- error,
) (common.Service, error) {
	listener, err := net.Listen("tcp", cfg.ListenEndpoint)
	if err != nil {
		return nil, err
	}

	subLogger := locator.Logger.With().Fields(map[string]interface{}{
		"subsystem": "frontend",
	}).Logger()

	s := &server{
		computer: locator.Computer,
		storage:  locator.Storage,
		logger:   &subLogger,
		errChan:  errChan,
		listener: listener,
		cfg:      cfg,
	}

	s.grpcServer = grpc.NewServer()

	schema.RegisterMemprofilerFrontendServer(s.grpcServer, s)
	reflection.Register(s.grpcServer)

	return s, nil
}
