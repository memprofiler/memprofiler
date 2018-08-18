package api

import (
	"net"

	"github.com/sirupsen/logrus"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/locator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var _ Service = (*server)(nil)

type server struct {
	grpcServer      *grpc.Server
	listener        net.Listener
	protocolFactory protocolFactory
	logger          logrus.FieldLogger
	errChan         chan<- error
	cfg             *config.APIConfig
}

func (s *server) Start() {
	s.errChan <- s.grpcServer.Serve(s.listener)
}

func (s *server) Stop() {
	s.grpcServer.GracefulStop()
}

func (s *server) Save(stream schema.Memprofiler_SaveServer) error {
	s.logger.Debug("Started request handling")

	// create object that will be responsible for handling incoming messages
	protocol := s.protocolFactory.save()
	defer func() {
		if err := protocol.close(); err != nil {
			s.logger.WithError(err).Error("Failed to close save protocol")
		}
	}()

	for {
		request, err := stream.Recv()
		if err != nil {
			return err
		}
		switch request.Payload.(type) {
		case *schema.SaveRequest_ServiceDescription:
			err = protocol.addDescription(request.GetServiceDescription())
		case *schema.SaveRequest_Measurement:
			err = protocol.addMeasurement(request.GetMeasurement())
		}
		if err != nil {
			s.logger.WithError(err).Error("Save error")
			return err
		}
	}
}

// NewServer builds new GRPC server
func NewServer(
	cfg *config.APIConfig,
	locator *locator.Locator,
	errChan chan<- error,
) (Service, error) {

	listener, err := net.Listen("tcp", cfg.ListenEnpdoint)
	if err != nil {
		return nil, err
	}

	s := &server{
		protocolFactory: &defaultProtocolFactory{locator: locator},
		listener:        listener,
		cfg:             cfg,
		errChan:         errChan,
		logger:          locator.Logger,
	}

	s.grpcServer = grpc.NewServer()
	schema.RegisterMemprofilerServer(s.grpcServer, s)
	reflection.Register(s.grpcServer)

	return s, nil
}
