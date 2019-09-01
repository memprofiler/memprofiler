package test

import (
	"context"
	"go/build"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"

	"github.com/memprofiler/memprofiler/schema"
)

func TestIntegration(t *testing.T) {
	var (
		projectPath             = filepath.Join(build.Default.GOPATH, "src/github.com/memprofiler/memprofiler")
		serverCfgPathFilesystem = filepath.Join(projectPath, "server/config/example_filesystem.yml")
		serverCfgPathTSDB       = filepath.Join(projectPath, "server/config/example_tsdb.yml")
	)
	t.Run("filesystemStorage", testTemplate(projectPath, serverCfgPathFilesystem))
	t.Run("tsdbStorage", testTemplate(projectPath, serverCfgPathTSDB))
}

func testTemplate(projectPath, serverConfigPath string) func(t *testing.T) {
	return func(t *testing.T) {
		testStartTime := time.Now()

		// wait to make 1 second pass, because session start times are aligned by seconds
		time.Sleep(1 * time.Second)

		// run environment
		e, err := newEnv(projectPath, serverConfigPath)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		defer e.Stop()

		// wait until client will send couple of reports to server
		time.Sleep(2 * e.reporterCfg.Scenario.Steps[0].Wait.Duration)

		expectedServiceName := e.reporterCfg.Client.InstanceDescription.ServiceName
		expectedInstanceName := e.reporterCfg.Client.InstanceDescription.InstanceName

		// 1. ask for list of service types
		getServicesRequest := &schema.GetServicesRequest{}
		getServicesResponse, err := e.client.GetServices(context.Background(), getServicesRequest)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		// there should be the only service right now
		assert.Len(t, getServicesResponse.Services, 1)
		assert.Equal(t, expectedServiceName, getServicesResponse.Services[0])

		// 2. ask for list of service instances
		getInstancesRequest := &schema.GetInstancesRequest{Service: expectedServiceName}
		getInstancesResponse, err := e.client.GetInstances(context.Background(), getInstancesRequest)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		// there should be the only instance right now
		assert.Len(t, getInstancesResponse.Instances, 1)
		assert.Equal(t, expectedServiceName, getInstancesResponse.Instances[0].ServiceName)
		assert.Equal(t, expectedInstanceName, getInstancesResponse.Instances[0].InstanceName)

		// 3. ask for list of sessions
		getSessionsRequest := &schema.GetSessionsRequest{
			Instance: &schema.InstanceDescription{
				ServiceName:  expectedServiceName,
				InstanceName: expectedInstanceName,
			},
		}
		getSessionResponse, err := e.client.GetSessions(context.Background(), getSessionsRequest)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		// there should be the only session right now
		assert.Len(t, getSessionResponse.Sessions, 1)
		session := getSessionResponse.Sessions[0]
		assert.Equal(t, expectedServiceName, session.Description.InstanceDescription.ServiceName)
		assert.Equal(t, expectedInstanceName, session.Description.InstanceDescription.InstanceName)
		assert.Equal(t, int64(1), session.Description.Id)

		// session start time must be greater than test start time
		sessionStartTime, err := ptypes.Timestamp(session.Metadata.StartedAt)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		assert.True(
			t,
			testStartTime.Before(sessionStartTime),
			"testStartTime=%v sessionStartTime=%v",
			testStartTime, sessionStartTime,
		)

		// session is still operational, so the finish time is nil
		assert.Nil(t, session.Metadata.FinishedAt)

		// 4. subscribe for session updates
		subscriptionRequest := &schema.SubscribeForSessionRequest{
			Session: session.Description,
		}
		subscription, err := e.client.SubscribeForSession(context.Background(), subscriptionRequest)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		// try to get a message from subscription and validate it
		metrics, err := subscription.Recv()
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		assert.NotNil(t, metrics)
		err = subscription.CloseSend()
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}
}
