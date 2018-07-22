package api

import (
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/common"
)

type Service interface {
	schema.MemprofilerServer
	common.Service
}
