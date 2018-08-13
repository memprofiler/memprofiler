package client

// Logger used within memprofiler client package
type Logger interface {
	Debug(string)
	Warning(string)
	Error(string)
}
