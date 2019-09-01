package common

// Service some significant part of Memprofiler server that can be started an stopped, than restarted
type Service interface {
	Start()
	Stop()
}

// Subsystem is something that can be gracefully stopped
type Subsystem interface {
	Quit()
}
