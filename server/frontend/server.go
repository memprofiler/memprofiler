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
	"github.com/memprofiler/memprofiler/server/storage/data"
	"github.com/memprofiler/memprofiler/server/storage/metadata"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var _ schema.MemprofilerFrontendServer = (*server)(nil)

type server struct {
	// httpServer *http.Server
	grpcServer      *grpc.Server
	listener        net.Listener
	computer        metrics.Computer
	dataStorage     data.Storage
	metadataStorage metadata.Storage
	errChan         chan<- error
	logger          *zerolog.Logger
}

func (s *server) GetServices(
	ctx context.Context,
	request *schema.GetServicesRequest,
) (*schema.GetServicesResponse, error) {
	result, err := s.metadataStorage.GetServices(ctx)
	if err != nil {
		return nil, err
	}
	return &schema.GetServicesResponse{Services: result}, nil
}

func (s *server) GetInstances(
	ctx context.Context,
	request *schema.GetInstancesRequest,
) (*schema.GetInstancesResponse, error) {
	instances, err := s.metadataStorage.GetInstances(ctx, request.GetService())
	if err != nil {
		// TODO: think about google.golang.org/grpc/status
		return nil, err
	}
	return &schema.GetInstancesResponse{Instances: instances}, nil
}

func (s *server) GetSessions(
	ctx context.Context,
	request *schema.GetSessionsRequest,
) (*schema.GetSessionsResponse, error) {
	sessions, err := s.metadataStorage.GetSessions(ctx, request.GetInstance())
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
	subscription, err := s.computer.SessionSubscribe(stream.Context(), request.GetSession())
	if subscription != nil {
		defer subscription.Unsubscribe()
	}
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to subscribe for session")
		return err
	}

	// push session metrics to the client
	for {
		select {
		case msg, ok := <-subscription.Updates():
			if !ok {
				// session terminated by a service
				s.logger.Warn().Msg("Session terminated by subscription broker")
				return nil
			}
			// sort trend values by InUseBytes rate, since it the most relevant indicator for memory leak
			sort.Slice(msg.Locations, func(i, j int) bool {
				// descending order
				return msg.Locations[i].Rates[0].Values.InUseBytes > msg.Locations[j].Rates[0].Values.InUseBytes
			})
			if err := stream.Send(msg); err != nil {
				s.logger.Error().Err(err).Msg("Failed to send msg to stream")
				return err
			}
		case <-stream.Context().Done():
			s.logger.Warn().Err(stream.Context().Err()).Msg("Context done")
			return nil
		}
	}
}

func (s *server) Start() { s.errChan <- s.grpcServer.Serve(s.listener) }

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
		computer:        locator.Computer,
		dataStorage:     locator.DataStorage,
		metadataStorage: locator.MetadataStorage,
		logger:          &subLogger,
		errChan:         errChan,
		listener:        listener,
	}

	s.grpcServer = grpc.NewServer()
	schema.RegisterMemprofilerFrontendServer(s.grpcServer, s)
	reflection.Register(s.grpcServer)

	return s, nil
}
