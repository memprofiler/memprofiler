package backend

import (
	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/common"
)

// Service provide functionality of GRPC server
type Service interface {
	schema.MemprofilerBackendServer
	common.Service
}
