package filesystem

import (
	"testing"

	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

var (
	stubServiceDescription = &schema.ServiceDescription{
		Type:     "crab",
		Instance: "local",
	}
	stubMMMeta = &measurementMetadata{
		SessionDescription: storage.SessionDescription{
			ServiceDescription: stubServiceDescription,
			SessionID:          1,
		},
		mmID: 1,
	}
	stubMM = &schema.Measurement{
		ObservedAt: &timestamp.Timestamp{
			Seconds: 1,
			Nanos:   1,
		},
		Locations: []*schema.Location{
			{
				MemoryUsage: &schema.MemoryUsage{AllocObjects: 1, AllocBytes: 2, FreeObjects: 3, FreeBytes: 4},
				CallStack: &schema.CallStack{
					Frames: []*schema.StackFrame{
						{
							Name: "a",
							File: "b",
							Line: 3,
						},
					},
				},
			},
		},
	}
)

func TestCache_SimpleOperations(t *testing.T) {
	cfg := config.FilesystemStorageCacheConfig{
		MaxTotalSize: 128,
		TTL:          10,
		GCFrequency:  1,
		GCMaxPause:   1,
	}
	cache := newCache(cfg)

	ok := cache.put(stubMMMeta, stubMM)
	assert.True(t, ok)
	result, ok := cache.get(stubMMMeta)
	assert.True(t, ok)
	assert.Equal(t, stubMM, result)

	cache.quit()
	assert.Equal(t, proto.Size(stubMM), cache.(*boundedLRUCache).actualSize)
}

func TestCache_Expiration(t *testing.T) {
	cfg := config.FilesystemStorageCacheConfig{
		MaxTotalSize: 128,
		TTL:          1,
		GCFrequency:  2,
		GCMaxPause:   1,
	}
	cache := newCache(cfg)

	// Read your writes
	ok := cache.put(stubMMMeta, stubMM)
	assert.True(t, ok)

	result, ok := cache.get(stubMMMeta)
	assert.True(t, ok)
	assert.Equal(t, stubMM, result)

	// Wait for GC, check that outdated message has been collected
	time.Sleep(time.Second * time.Duration(cfg.GCFrequency) * 2)
	result, ok = cache.get(stubMMMeta)
	assert.False(t, ok)
	assert.Nil(t, result)

	cache.quit()
	assert.Equal(t, 0, cache.(*boundedLRUCache).actualSize)
}
