package common

type Service interface {
	Start()
	Stop()
}

type Subsystem interface {
	Quit()
}
