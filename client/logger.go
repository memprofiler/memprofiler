package client

// Logger that is used within memprofiler client package
type Logger interface {
	Debug(msg string)
	Warning(msg string)
	Error(msg string)
}
