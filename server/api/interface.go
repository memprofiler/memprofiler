package api

import (
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/common"
)

// Service provide functionality of GRPC server
type Service interface {
	schema.MemprofilerServer
	common.Service
}
