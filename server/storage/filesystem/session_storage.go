package filesystem

import (
	"fmt"
	"sort"
	"sync"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage"
)

type sessionStorage interface {
	inc(*schema.ServiceDescription) storage.SessionID
	register(*schema.ServiceDescription, storage.SessionID)
	storage.MetadataStorage
}

var _ sessionStorage = (*defaultSessionStorage)(nil)

// defaultSessionStorage stores SessionID values in memory
type defaultSessionStorage struct {
	mutex sync.RWMutex
	// ServiceType <-> ServiceInstance <-> []SessionID
	values map[string]map[string][]storage.SessionID
}

func (ss *defaultSessionStorage) inc(desc *schema.ServiceDescription) storage.SessionID {

	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	instances, exists := ss.values[desc.GetServiceType()]
	if !exists {
		instances = map[string][]storage.SessionID{}
		ss.values[desc.GetServiceInstance()] = instances
	}

	sessionIDs, exists := instances[desc.GetServiceInstance()]
	if !exists {
		value := storage.SessionID(0)
		instances[desc.GetServiceInstance()] = []storage.SessionID{value}
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

	instances, exists := ss.values[desc.GetServiceType()]
	if !exists {
		instances = map[string][]storage.SessionID{}
		ss.values[desc.GetServiceType()] = instances
	}

	sessionIDs, exists := instances[desc.GetServiceInstance()]
	if !exists {
		instances[desc.GetServiceInstance()] = []storage.SessionID{value}
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

func (ss *defaultSessionStorage) Instances(serviceType string) ([]*schema.ServiceDescription, error) {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	instances, exists := ss.values[serviceType]
	if !exists {
		return nil, fmt.Errorf("no services of type '%s' are registered", serviceType)
	}
	results := make([]*schema.ServiceDescription, 0, len(instances))
	for instance := range instances {
		results = append(results, &schema.ServiceDescription{ServiceType: serviceType, ServiceInstance: instance})
	}
	return results, nil
}

func (ss *defaultSessionStorage) Sessions(desc *schema.ServiceDescription) ([]*schema.Session, error) {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	instances, exists := ss.values[desc.GetServiceType()]
	if !exists {
		return nil, fmt.Errorf("no services of type '%s' are registered", desc.GetServiceType())
	}

	sessionIDs, exists := instances[desc.GetServiceInstance()]
	if !exists {
		return nil, fmt.Errorf("no sessions for service '%s' of type '%s' are registered", desc.GetServiceInstance(), desc.GetServiceType())
	}

	results := make([]*schema.Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session := &schema.Session{
			Description: &schema.SessionDescription{
				ServiceType:     desc.GetServiceType(),
				ServiceInstance: desc.GetServiceInstance(),
				SessionId:       sessionID,
			},
		}
		results = append(results, session)
	}
	return results, nil
}

func newSessionStorage() sessionStorage {
	return &defaultSessionStorage{
		mutex:  sync.RWMutex{},
		values: make(map[string]map[string][]storage.SessionID),
	}
}
