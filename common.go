package analyzer

import (
	"crypto/md5"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
)

// Measurement provides an instantaneous memory usage stats
type Measurement struct {
	Timestamp int64       `json:"timestamp"`
	Locations []*Location `json:"locations"`
}

// Location describes memory stats and the code where it was allocated
type Location struct {
	Stack       *Stack       `json:"stack"`
	MemoryUsage *MemoryUsage `json:"memory_usage"`
}

// MemoryUsage keeps memory profiler records
type MemoryUsage struct {
	InUseObjects int64 `json:"in_use_objects"`
	InUseBytes   int64 `json:"in_use_bytes"`
	AllocObjects int64 `json:"alloc_objects"`
	AllocBytes   int64 `json:"alloc_bytes"`
}

// Stack describes the call stack of memory allocations
type Stack struct {
	Records []*StackRecord
}

// fill uses raw data to populate stack
func (s *Stack) fill(raw []uintptr, allFrames bool) {
	var (
		show   = allFrames
		frames = runtime.CallersFrames(raw)
	)
	for {
		frame, more := frames.Next()
		name := frame.Function
		if name == "" {
			show = true
		} else if name != "runtime.goexit" && (show || !strings.HasPrefix(name, "runtime.")) {
			show = true
			record := &StackRecord{Name: name, File: frame.File, Line: frame.Line}
			s.Records = append(s.Records, record)
		}
		if !more {
			break
		}
	}

	if !show {
		s.fill(rawStack, true)
	}
}

// hash helps to compare stacks
func (s *Stack) hash() (string, error) {
	h := md5.New()

	for _, sr := range s.Records {
		if _, err := io.WriteString(h, sr.dump()); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// StackRecord describes the frame of call stack
type StackRecord struct {
	Name string `json:"name"`
	File string `json:"file"`
	Line int    `json:"line"`
}

func (sr *StackRecord) dump() string {
	return sr.Name + sr.File + strconv.Itoa(sr.Line)
}
