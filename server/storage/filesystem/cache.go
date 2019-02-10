package filesystem

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vitalyisaev2/memprofiler/server/storage"

	"github.com/golang/protobuf/proto"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/config"
)

type measurementID uint

type measurementMetadata struct {
	session *storage.SessionDescription
	mmID    measurementID
}

type measurementCached struct {
	mm        *schema.Measurement // original measurement
	createdAt int64               // seconds since the beginning of the epoch
}

func (meta *measurementMetadata) String() string {
	return fmt.Sprintf("%s::%d", meta.session.String(), meta.mmID)
}

type cache interface {
	put(meta *measurementMetadata, mm *schema.Measurement) bool
	get(meta *measurementMetadata) (*schema.Measurement, bool)
	quit()
}

// boundedLRUCache is a simple implementation of cache providing:
// 1. thread-safety
// 2. limitation for sum of values sizes;
// 3. deletion of values with expired TTL;
// TODO: sharding? + learn caching strategies
type boundedLRUCache struct {
	values     map[string]*measurementCached // cached values
	mutex      sync.RWMutex                  // synchronizes access to cache
	actualSize int                           // current sum of value sizes

	cfg    config.FilesystemStorageCacheConfig
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *boundedLRUCache) put(meta *measurementMetadata, mm *schema.Measurement) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// don't exceed memory limitation
	mmSize := proto.Size(mm)
	if c.actualSize+mmSize > c.cfg.MaxTotalSize {
		return false
	}

	// put value to cache (maybe override old record)
	c.values[meta.String()] = &measurementCached{
		mm:        mm,
		createdAt: time.Now().Unix(),
	}
	c.actualSize += mmSize
	return true
}

func (c *boundedLRUCache) get(meta *measurementMetadata) (*schema.Measurement, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, exists := c.values[meta.String()]
	if !exists {
		return nil, false
	}
	return value.mm, true
}

func (c *boundedLRUCache) loop() {
	defer c.wg.Done()
	duration := time.Duration(c.cfg.GCFrequency) * time.Second
	ticker := time.Tick(duration)
	for {
		select {
		case <-ticker:
			c.collectGarbage()
		case <-c.ctx.Done():
			return
		}
	}
}

// collectGarbage simply iterates through the map and deletes outdated messages
func (c *boundedLRUCache) collectGarbage() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// compute timestamp that will be a limit for outdated messages
	ttl := time.Duration(-1*c.cfg.TTL) * time.Second
	outdated := time.Now().Add(ttl).Unix()

	// create timer to control GC's sessionDescription duration
	maxGCPause := time.Duration(c.cfg.GCMaxPause) * time.Second
	timer := time.NewTimer(maxGCPause)
	for k, v := range c.values {

		// check for timeout, in order to not to block cache for too long
		select {
		case <-timer.C:
			return
		case <-c.ctx.Done():
			return
		default:
		}

		// delete outdated message
		if v.createdAt < outdated {
			delete(c.values, k)
			c.actualSize -= proto.Size(v.mm)
		}
	}
}

func (c *boundedLRUCache) quit() {
	c.cancel()
	c.wg.Wait()
}

func newCache(cfg config.FilesystemStorageCacheConfig) cache {
	ctx, cancel := context.WithCancel(context.Background())
	c := &boundedLRUCache{
		values: make(map[string]*measurementCached),
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
	}
	c.wg.Add(1)
	go c.loop()
	return c
}
