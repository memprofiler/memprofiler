package backend

import "github.com/memprofiler/memprofiler/server/locator"

type protocolFactory interface {
	save() saveProtocol
}

type defaultProtocolFactory struct {
	locator *locator.Locator
}

func (f *defaultProtocolFactory) save() saveProtocol {
	return newSaveProtocol(f.locator)
}
