package test

import (
	"context"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/storage"
	"github.com/memprofiler/memprofiler/server/storage/filesystem"
	"github.com/memprofiler/memprofiler/server/storage/tsdb"
)

// TestStorageWriteReadSimpleLocations simple integration test for tsdb-based storage for simple locations
func TestStorageWriteReadSimpleLocations(t *testing.T) {
	input := []*schema.Measurement{
		{
			ObservedAt: &timestamp.Timestamp{Seconds: 1},
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 1},
					Callstack: &schema.Callstack{
						Id:     "abcd",
						Frames: []*schema.StackFrame{{Name: "a", File: "b.go", Line: 1}},
					},
				},
			},
		},
		{
			ObservedAt: &timestamp.Timestamp{Seconds: 2},
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 2},
					Callstack: &schema.Callstack{
						Id:     "edfg",
						Frames: []*schema.StackFrame{{Name: "b", File: "c.go", Line: 2}},
					},
				},
			},
		},
	}

	// for this case input == output because we have one location for second
	t.Run("filesystem", caseStorageWriteRead(newStorage(t, false), input, input))
	t.Run("tsdb", caseStorageWriteRead(newStorage(t, true), input, input))
}

// TestStorageWriteReadManyLocations simple integration test for tsdb-based storage
func TestStorageWriteReadManyLocations(t *testing.T) {
	input := []*schema.Measurement{
		{
			ObservedAt: &timestamp.Timestamp{Seconds: 1},
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 1},
					Callstack: &schema.Callstack{
						Id:     "abcd",
						Frames: []*schema.StackFrame{{Name: "a", File: "b.go", Line: 1}},
					},
				},
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 11},
					Callstack: &schema.Callstack{
						Id:     "edfg",
						Frames: []*schema.StackFrame{{Name: "b", File: "b.go", Line: 2}},
					},
				},
			},
		},
		{
			ObservedAt: &timestamp.Timestamp{Seconds: 1},
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 111},
					Callstack: &schema.Callstack{
						Id:     "hijk",
						Frames: []*schema.StackFrame{{Name: "c", File: "c.go", Line: 3}},
					},
				},
			},
		},
		{
			ObservedAt: &timestamp.Timestamp{Seconds: 2},
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 2},
					Callstack: &schema.Callstack{
						Id:     "lmno",
						Frames: []*schema.StackFrame{{Name: "d", File: "d.go", Line: 4}},
					},
				},
			},
		},
	}
	output := []*schema.Measurement{
		{
			ObservedAt: &timestamp.Timestamp{Seconds: 1},
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 1},
					Callstack: &schema.Callstack{
						Id:     "abcd",
						Frames: []*schema.StackFrame{{Name: "a", File: "b.go", Line: 1}},
					},
				},
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 11},
					Callstack: &schema.Callstack{
						Id:     "edfg",
						Frames: []*schema.StackFrame{{Name: "b", File: "b.go", Line: 2}},
					},
				},
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 111},
					Callstack: &schema.Callstack{
						Id:     "hijk",
						Frames: []*schema.StackFrame{{Name: "c", File: "c.go", Line: 3}},
					},
				},
			},
		},
		{
			ObservedAt: &timestamp.Timestamp{Seconds: 2},
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 2},
					Callstack: &schema.Callstack{
						Id:     "lmno",
						Frames: []*schema.StackFrame{{Name: "d", File: "d.go", Line: 4}},
					},
				},
			},
		},
	}

	// input != output, because we write in one moment separately two different locations
	// TODO: does not work, need research
	//t.Run("filesystem", caseStorageWriteRead(newStorage(t, false), input, output))
	t.Run("tsdb", caseStorageWriteRead(newStorage(t, true), input, output))
}

func caseStorageWriteRead(s storage.Storage, input, expected []*schema.Measurement) func(t *testing.T) {
	return func(t *testing.T) {
		// write some measurements
		serviceDesc := &schema.ServiceDescription{
			ServiceType:     "database",
			ServiceInstance: "localhost:8080",
		}
		saver, err := s.NewDataSaver(serviceDesc)
		assert.NoError(t, err)
		assert.NotNil(t, saver)

		for _, mm := range input {
			err = saver.Save(mm)
			assert.NoError(t, err)
		}
		err = saver.Close()
		assert.NoError(t, err)

		// try to load data just written
		sessionDesc := &schema.SessionDescription{
			ServiceType:     serviceDesc.GetServiceType(),
			ServiceInstance: serviceDesc.GetServiceInstance(),
			SessionId:       saver.SessionID(),
		}
		loader, err := s.NewDataLoader(sessionDesc)
		assert.NotNil(t, loader)
		assert.NoError(t, err)

		outChan, err := loader.Load(context.Background())
		assert.NotNil(t, outChan)
		assert.NoError(t, err)

		output := make([]*schema.Measurement, 0, len(expected))
		for result := range outChan {
			assert.NotNil(t, result.Measurement)
			if !assert.NoError(t, result.Err) {
				assert.FailNow(t, "failed to read data")
			}
			output = append(output, result.Measurement)
		}

		err = loader.Close()
		assert.NoError(t, err)

		if !assert.Equal(t, len(expected), len(output)) {
			return
		}
		for k, expectedMeasurement := range expected {
			currentMeasurement := output[k]
			assert.True(t, compareLocationsSets(expectedMeasurement.Locations, currentMeasurement.Locations))
		}
	}
}

func compareLocationsSets(l1, l2 []*schema.Location) bool {
	if len(l1) != len(l2) {
		return false
	}
	for _, i := range l1 {
		res := false
		for _, j := range l2 {
			if reflect.DeepEqual(i.GetCallstack(), j.GetCallstack()) && reflect.DeepEqual(i.GetMemoryUsage(), j.GetMemoryUsage()) {
				res = true
				break
			}
		}
		if !res {
			return false
		}
	}

	return true
}

func newStorage(t *testing.T, isTSDB bool) storage.Storage {
	logger := logrus.New()
	logger.Out = os.Stdout

	// create new storage in tmp dir
	dataDir, err := ioutil.TempDir("/tmp", "memprofiler")
	assert.NoError(t, err)

	defer func() {
		if err = os.RemoveAll(dataDir); err != nil {
			logger.WithError(err).Fatal("Failed to remove dir")
		}
	}()

	var s storage.Storage
	if isTSDB {
		cfg := &config.TSDBStorageConfig{
			DataDir: dataDir,
		}
		s, err = tsdb.NewStorage(logger, cfg)
	} else {
		cfg := &config.FilesystemStorageConfig{
			DataDir:   dataDir,
			SyncWrite: false,
		}
		s, err = filesystem.NewStorage(logger, cfg)
	}
	assert.NotNil(t, s)
	assert.NoError(t, err)

	return s
}
