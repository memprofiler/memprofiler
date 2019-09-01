package metadata

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/utils"
)

func TestStorage(t *testing.T) {
	logger := utils.NewLogger(&config.LoggingConfig{Level: zerolog.DebugLevel})

	// create new storage instance
	dirName, err := ioutil.TempDir("", "memprofiler")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(dirName)

	storage, err := NewStorageSQLite(logger, &config.MetadataStorageConfig{DataDir: dirName})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer storage.Quit()

	// there are two services with one instance per service
	ctx := context.Background()
	instances := []*schema.InstanceDescription{
		{ServiceName: "service1", InstanceName: "localhost:12345"},
		{ServiceName: "service2", InstanceName: "www.ya.ru:80"},
	}

	expectedStartTime := time.Now()

	t.Run("StartSession", func(t *testing.T) {
		// services and instances are created by demand, when the new session is created
		for i, instanceDesc := range instances {
			sessionDesc, err := storage.StartSession(ctx, instanceDesc)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, int64(i+1), sessionDesc.Id)
			assert.Equal(t, instanceDesc.InstanceName, sessionDesc.Instance.InstanceName)
			assert.Equal(t, instanceDesc.ServiceName, sessionDesc.Instance.ServiceName)
		}
	})

	t.Run("GetServices", func(t *testing.T) {
		// get service list
		actualServices, err := storage.GetServices(ctx)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, actualServices, len(instances))
		assert.Equal(t, actualServices[0], instances[0].ServiceName)
		assert.Equal(t, actualServices[1], instances[1].ServiceName)
	})

	t.Run("GetInstances", func(t *testing.T) {
		// get instance list (x2)
		actualInstances, err := storage.GetInstances(ctx, instances[0].ServiceName)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, actualInstances, 1)
		assert.Equal(t, actualInstances[0], instances[0])

		actualInstances, err = storage.GetInstances(ctx, instances[1].ServiceName)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, actualInstances, 1)
		assert.Equal(t, actualInstances[0], instances[1])
	})

	t.Run("GetSessions", func(t *testing.T) {
		for i, instanceDesc := range instances {
			sessions, err := storage.GetSessions(ctx, instanceDesc)
			if !assert.NoError(t, err) {
				return
			}
			assert.Len(t, sessions, 1)

			session := sessions[0]

			assert.Equal(t, instanceDesc, session.Description.Instance)
			assert.Equal(t, int64(i+1), session.Description.Id)

			actualStartTime, err := ptypes.Timestamp(session.Metadata.StartedAt)
			assert.NoError(t, err)

			assert.True(t, expectedStartTime.Add(time.Second).After(actualStartTime)) // started recently, about 1 sec ago
			assert.Nil(t, session.Metadata.FinishedAt)                                // yet not stopped
		}
	})

	t.Run("StopSessions", func(t *testing.T) {
		for _, instanceDesc := range instances {
			sessionsBeforeStop, err := storage.GetSessions(ctx, instanceDesc)
			assert.NoError(t, err)

			session := sessionsBeforeStop[0]

			err = storage.StopSession(ctx, session.Description)
			assert.NoError(t, err)

			expectedFinishTime := time.Now()

			sessionsAfterStop, err := storage.GetSessions(ctx, instanceDesc)
			assert.NoError(t, err)

			session = sessionsAfterStop[0]

			actualFinishTime, err := ptypes.Timestamp(session.Metadata.FinishedAt)
			assert.NoError(t, err)

			assert.True(t, expectedFinishTime.Add(time.Second).After(actualFinishTime)) // session just stopped
		}
	})
}
