package filesystem

import (
	"sort"
	"sync"

	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

type sessionStorage interface {
	inc(*schema.ServiceDescription) storage.SessionID
	register(*schema.ServiceDescription, storage.SessionID)
	storage.MetadataStorage
}

var _ sessionStorage = (*defaultSessionStorage)(nil)

// defaultSessionStorage stores SessionID values
type defaultSessionStorage struct {
	mutex sync.RWMutex
	// ServiceID <-> InstanceID <-> []SessionID
	values map[string]map[string][]storage.SessionID
}

func (ss *defaultSessionStorage) inc(desc *schema.ServiceDescription) storage.SessionID {

	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	instances, exists := ss.values[desc.GetType()]
	if !exists {
		instances = map[string][]storage.SessionID{}
		ss.values[desc.GetType()] = instances
	}

	sessionIDs, exists := instances[desc.GetInstance()]
	if !exists {
		value := storage.SessionID(0)
		instances[desc.GetInstance()] = []storage.SessionID{value}
		return value
	}

	newSesssionID := sessionIDs[len(sessionIDs)-1] + 1
	sessionIDs = append(sessionIDs, newSesssionID)
	sort.Slice(sessionIDs, func(i, j int) bool { return sessionIDs[i] < sessionIDs[j] })
	return newSesssionID
}

func (ss *defaultSessionStorage) register(
	desc *schema.ServiceDescription,
	value storage.SessionID) {

	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	instances, exists := ss.values[desc.GetType()]
	if !exists {
		instances = map[string][]storage.SessionID{}
		ss.values[desc.GetType()] = instances
	}

	sessionIDs, exists := instances[desc.GetInstance()]
	if !exists {
		instances[desc.GetInstance()] = []storage.SessionID{value}
		return
	}

	sessionIDs = append(sessionIDs, value)
	sort.Slice(sessionIDs, func(i, j int) bool { return sessionIDs[i] < sessionIDs[j] })
}

func (ss *defaultSessionStorage) Services() []string {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	results := make([]string, len(ss.values))
	i := 0
	for serviceType := range ss.values {
		results[i] = serviceType
		i++
	}
	return results
}

func (ss *defaultSessionStorage) Instances(serviceType string) []string {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	instances, exists := ss.values[serviceType]
	if !exists {
		return nil
	}
	results := make([]string, len(instances))
	i := 0
	for instance := range instances {
		results[i] = instance
		i++
	}
	return results
}

func (ss *defaultSessionStorage) Sessions(desc *schema.ServiceDescription) []storage.SessionID {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	instances, exists := ss.values[desc.GetType()]
	if !exists {
		return nil
	}

	sessionIDs, exists := instances[desc.GetInstance()]
	if !exists {
		return nil
	}

	result := make([]storage.SessionID, len(sessionIDs))
	copy(result, sessionIDs)
	return result
}

func newSessionStorage() sessionStorage {
	return &defaultSessionStorage{
		mutex:  sync.RWMutex{},
		values: make(map[string]map[string][]storage.SessionID),
	}
}
