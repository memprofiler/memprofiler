package test

import (
	"bytes"
	"os"
	"os/exec"
)

const (
	serverBinary = "memprofiler"
	dataDir      = "/tmp/memprofiler"
)

// serverLauncher runs memprofiler service in a distinct process
type serverLauncher struct {
	process *os.Process
	buf     *bytes.Buffer
}

func (l *serverLauncher) start() error {
	// clean data directory
	if err := l.cleanDataDir(); err != nil {
		return err
	}

	// launch server in a separate process
	cmd := exec.Command(serverBinary, "-c", "./server/config/example.yml")
	cmd.Stdout = l.buf
	cmd.Stderr = l.buf

	err := cmd.Start()
	l.process = cmd.Process
	return err
}

func (l *serverLauncher) cleanDataDir() error {
	_, err := os.Stat(dataDir)
	if os.IsExist(err) {
		if err := os.RemoveAll(dataDir); err != nil {
			return err
		}
	}
	return os.MkdirAll(dataDir, 0644)
}

func (l *serverLauncher) stop() error {
	if l.process != nil {
		return l.process.Kill()
	}
	return nil
}

//func (l *serverLauncher) logs() string { return l.buf.String() }

func newServerLauncher() *serverLauncher {
	return &serverLauncher{
		buf: bytes.NewBuffer(nil),
	}
}
