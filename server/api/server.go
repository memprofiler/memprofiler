package api

import (
	"net"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/locator"
	"google.golang.org/grpc"
)

var _ Service = (*server)(nil)

type server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	locator    *server.Locator
	errChan    chan<- error
	cfg        *config.Config
}

func (s *server) Start() {
	s.errChan <- s.grpcServer.Serve(s.listener)
}

func (s *server) Stop() {
	s.grpcServer.GracefulStop()
}

func (s *server) Save(stream schema.Memprofiler_SaveServer) error {
}

func NewService(
	cfg *config.ServerConfig,
	locator *locator.Locator,
	errChan chan<- error,
) (Service, error) {

	listener, err := net.Listen("tcp", cfg.Server.ListenEndpoint)
	if err != nil {
		return nil, err
	}

	s := &server{
		server:   grpc.NewServer(),
		locator:  locator,
		listener: listener,
		cfg:      cfg,
		errChan:  errChan,
	}

	return s, nil
}
