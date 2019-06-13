package tsdb

import (
	"fmt"
	"sort"
	"sync"

	"github.com/golang/protobuf/ptypes"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage"
)

type sessionStorage interface {
	storage.MetadataStorage
	registerNextSession(*schema.ServiceDescription) *schema.Session
	registerExistingSession(*schema.Session)
}

var _ sessionStorage = (*defaultSessionStorage)(nil)

// defaultSessionStorage stores session metadata in memory;
// the amount of memory consumed by sessions is not expected to be very big;
// TODO: good place for sqlite (?)
type defaultSessionStorage struct {
	mutex sync.RWMutex
	// ServiceType <-> ServiceInstance <-> []*Session
	values map[string]map[string][]*schema.Session
}

func (ss *defaultSessionStorage) registerNextSession(desc *schema.ServiceDescription) *schema.Session {

	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	session := &schema.Session{
		Description: &schema.SessionDescription{
			ServiceType:     desc.GetServiceType(),
			ServiceInstance: desc.GetServiceInstance(),
		},
		Metadata: &schema.SessionMetadata{
			StartedAt:  ptypes.TimestampNow(),
			FinishedAt: nil, // ongoing session
		},
	}

	instances, exists := ss.values[desc.GetServiceType()]
	if !exists {
		instances = map[string][]*schema.Session{}
		ss.values[desc.GetServiceInstance()] = instances
	}

	sessions, exists := instances[desc.GetServiceInstance()]
	if !exists {
		// case 1: this is the first session for this service instance
		session.Description.SessionId = 0
		instances[desc.GetServiceInstance()] = []*schema.Session{session}
	} else {
		// case 2: append session to list of existing sessions, increment counter
		session.Description.SessionId = sessions[len(sessions)-1].Description.SessionId + 1
		ss.values[desc.GetServiceType()][desc.GetServiceInstance()] = append(sessions, session)
	}

	return session
}

func (ss *defaultSessionStorage) registerExistingSession(session *schema.Session) {

	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	instances, exists := ss.values[session.GetDescription().GetServiceType()]
	if !exists {
		instances = map[string][]*schema.Session{}
		ss.values[session.GetDescription().GetServiceType()] = instances
	}

	sessions, exists := instances[session.GetDescription().GetServiceInstance()]
	if !exists {
		instances[session.GetDescription().GetServiceInstance()] = []*schema.Session{session}
		return
	}

	sessions = append(sessions, session)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].GetDescription().GetSessionId() < sessions[j].GetDescription().GetSessionId()
	})
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

	sessions, exists := instances[desc.GetServiceInstance()]
	if !exists {
		return nil, fmt.Errorf(
			"no sessions for service '%s' of type '%s' are registered",
			desc.GetServiceInstance(), desc.GetServiceType())
	}

	return sessions, nil
}

func newSessionStorage() sessionStorage {
	return &defaultSessionStorage{
		mutex:  sync.RWMutex{},
		values: make(map[string]map[string][]*schema.Session),
	}
}
