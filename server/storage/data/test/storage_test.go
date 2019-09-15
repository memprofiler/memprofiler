package test

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/storage/data"
	"github.com/memprofiler/memprofiler/server/storage/data/filesystem"
	"github.com/memprofiler/memprofiler/server/storage/data/tsdb"
	"github.com/memprofiler/memprofiler/server/storage/metadata"
	"github.com/memprofiler/memprofiler/utils"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestStorageWriteReadSimpleLocations simple integration test for tsdb-based storage for simple locations
func TestStorageWriteReadSimpleLocations(t *testing.T) {
	t.SkipNow()
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
	cases := []struct {
		name    string
		storage data.Storage
		input   []*schema.Measurement
		output  []*schema.Measurement
	}{
		{
			name:    "filesystem",
			storage: newStorage(t, config.FilesystemDataStorage),
			input:   input,
			output:  input,
		},
		{
			name:    "tsdb",
			storage: newStorage(t, config.TSDBDataStorage),
			input:   input,
			output:  input,
		},
	}
	for _, tc := range cases {
		if t.Failed() {
			return
		}
		t.Run(tc.name, caseStorageWriteRead(tc.storage, tc.input, tc.output))
	}
}

// TestStorageWriteReadManyLocations simple integration test for tsdb-based storage
func TestStorageWriteReadManyLocations(t *testing.T) {
	input := []*schema.Measurement{
		{
			ObservedAt: &timestamp.Timestamp{Seconds: 1},
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 1},
					Callstack:   &schema.Callstack{Id: "abcd", Frames: []*schema.StackFrame{{Name: "a", File: "b.go", Line: 1}}},
				},
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 11},
					Callstack:   &schema.Callstack{Id: "edfg", Frames: []*schema.StackFrame{{Name: "b", File: "b.go", Line: 2}}},
				},
			},
		},
		{
			ObservedAt: &timestamp.Timestamp{Seconds: 1},
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 111},
					Callstack:   &schema.Callstack{Id: "hijk", Frames: []*schema.StackFrame{{Name: "c", File: "c.go", Line: 3}}},
				},
			},
		},
		{
			ObservedAt: &timestamp.Timestamp{Seconds: 2},
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 2},
					Callstack:   &schema.Callstack{Id: "lmno", Frames: []*schema.StackFrame{{Name: "d", File: "d.go", Line: 4}}},
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
					Callstack:   &schema.Callstack{Id: "abcd", Frames: []*schema.StackFrame{{Name: "a", File: "b.go", Line: 1}}},
				},
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 11},
					Callstack:   &schema.Callstack{Id: "edfg", Frames: []*schema.StackFrame{{Name: "b", File: "b.go", Line: 2}}},
				},
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 111},
					Callstack:   &schema.Callstack{Id: "hijk", Frames: []*schema.StackFrame{{Name: "c", File: "c.go", Line: 3}}},
				},
			},
		},
		{
			ObservedAt: &timestamp.Timestamp{Seconds: 2},
			Locations: []*schema.Location{
				{
					MemoryUsage: &schema.MemoryUsage{AllocBytes: 2},
					Callstack:   &schema.Callstack{Id: "lmno", Frames: []*schema.StackFrame{{Name: "d", File: "d.go", Line: 4}}},
				},
			},
		},
	}

	// input != output, because we write in one moment separately two different locations
	cases := []struct {
		name    string
		storage data.Storage
		input   []*schema.Measurement
		output  []*schema.Measurement
	}{
		// В принципе не может проходить на FS хранилище
		// {name: "filesystem", storage: newStorage(t, config.FilesystemDataStorage), input: input, output: output},
		{name: "tsdb", storage: newStorage(t, config.TSDBDataStorage), input: input, output: output},
	}
	for _, tc := range cases {
		if t.Failed() {
			return
		}
		t.Run(tc.name, caseStorageWriteRead(tc.storage, tc.input, tc.output))
	}
}

func caseStorageWriteRead(s data.Storage, input, expected []*schema.Measurement) func(t *testing.T) {
	return func(t *testing.T) {
		// write some measurements
		instanceDesc := &schema.InstanceDescription{
			ServiceName:  "some_web_service",
			InstanceName: "localhost:8080",
		}
		saver, err := s.NewDataSaver(instanceDesc)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		assert.NotNil(t, saver)

		for _, mm := range input {
			err = saver.Save(mm)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
		}
		err = saver.Close()
		assert.NoError(t, err)

		// try to load data just written
		loader, err := s.NewDataLoader(saver.SessionDescription())
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

		if !assert.Equal(t, len(expected), len(output), "expected=%v actual=%v", expected, output) {
			t.FailNow()
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
			callstackEquality := reflect.DeepEqual(i.GetCallstack(), j.GetCallstack())
			memoryUsageEquality := reflect.DeepEqual(i.GetMemoryUsage(), j.GetMemoryUsage())
			if callstackEquality && memoryUsageEquality {
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

func newStorage(t *testing.T, dataStorageType config.DataStorageType) data.Storage {
	logger := utils.NewLogger(&config.LoggingConfig{Level: zerolog.DebugLevel})

	// create tmp dir
	rootDir, err := ioutil.TempDir("/tmp", "memprofiler")
	if !assert.NoError(t, err) {
		assert.FailNow(t, "failed to create tmp directory")
	}

	dataDir := filepath.Join(rootDir, "data")
	metadataDir := filepath.Join(rootDir, "metadata")

	// run metadata storage
	metadataStorageCfg := &config.MetadataStorageConfig{DataDir: metadataDir}
	metadataStorage, err := metadata.NewStorageSQLite(logger, metadataStorageCfg)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "failed to create metadata storage")
	}

	// run data storage
	var dataStorage data.Storage
	switch dataStorageType {
	case config.TSDBDataStorage:
		cfg := &config.TSDBStorageConfig{
			DataDir: dataDir,
		}
		dataStorage, err = tsdb.NewStorage(logger, cfg, metadataStorage)
	case config.FilesystemDataStorage:
		cfg := &config.FilesystemStorageConfig{
			DataDir:   dataDir,
			SyncWrite: false,
		}
		dataStorage, err = filesystem.NewStorage(logger, cfg, metadataStorage)
	}
	if !assert.NoError(t, err) {
		assert.FailNow(t, "failed to run data storage")
	}
	assert.NotNil(t, dataStorage)

	return dataStorage
}
