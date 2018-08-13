package api

import (
	"net"

	"github.com/sirupsen/logrus"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/locator"
	"google.golang.org/grpc"
)

var _ Service = (*server)(nil)

type server struct {
	grpcServer      *grpc.Server
	listener        net.Listener
	protocolFactory protocolFactory
	logger          *logrus.Logger
	errChan         chan<- error
	cfg             *config.ServerConfig
}

func (s *server) Start() {
	s.errChan <- s.grpcServer.Serve(s.listener)
}

func (s *server) Stop() {
	s.grpcServer.GracefulStop()
}

func (s *server) Save(stream schema.Memprofiler_SaveServer) error {

	// create object that will be responsible for storing measurements
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

// NewAPI builds new GRPC server
func NewAPI(
	cfg *config.ServerConfig,
	locator *locator.Locator,
	errChan chan<- error,
) (Service, error) {

	locator.Logger.WithField("endpoint", cfg.ListenEnpdoint).Debug("Starting listening")

	listener, err := net.Listen("tcp", cfg.ListenEnpdoint)
	if err != nil {
		return nil, err
	}

	s := &server{
		grpcServer:      grpc.NewServer(),
		protocolFactory: &defaultProtocolFactory{locator: locator},
		listener:        listener,
		cfg:             cfg,
		errChan:         errChan,
	}

	return s, nil
}
