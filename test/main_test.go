package test

import (
	"context"
	"go/build"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/config"
)

func TestIntegration(t *testing.T) {
	var (
		projectPath             = filepath.Join(build.Default.GOPATH, "src/github.com/memprofiler/memprofiler")
		serverCfgPathFilesystem = filepath.Join(projectPath, "server/config/example.yml")
		serverCfgPathTSDB       = filepath.Join(projectPath, "server/config/example_tsdb.yml")
	)
	t.Run("filesystemStorage", testTemplate(projectPath, serverCfgPathFilesystem))
	t.Run("tsdbStorage", testTemplate(projectPath, serverCfgPathTSDB))
}

func testTemplate(projectPath, serverConfigPath string) func(t *testing.T) {
	return func(t *testing.T) {
		testStartTime := time.Now()

		// run environment
		e, err := newEnv(projectPath, serverConfigPath)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Stop()

		// wait until client will send couple of reports to server
		time.Sleep(2 * e.reporterCfg.Scenario.Steps[0].Wait.Duration)

		expectedServiceType := e.reporterCfg.Client.ServiceDescription.ServiceType
		expectedServiceInstance := e.reporterCfg.Client.ServiceDescription.ServiceInstance

		// 1. ask for list of service types
		getServicesRequest := &schema.GetServicesRequest{}
		getServicesResponse, err := e.client.GetServices(context.Background(), getServicesRequest)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		// there should be the only service right now
		assert.Len(t, getServicesResponse.ServiceTypes, 1)
		assert.Equal(t, expectedServiceType, getServicesResponse.ServiceTypes[0])

		// 1.1 validate http
		data, err := sendHTTPReq(e.serverCfg, "services")
		require.NoError(t, err)

		httpGetServicesResponse := &schema.GetServicesResponse{}
		err = jsonpb.Unmarshal(data, httpGetServicesResponse)
		require.NoError(t, err)
		require.True(t, proto.Equal(httpGetServicesResponse, getServicesResponse))

		// 2. ask for list of service instances
		getInstancesRequest := &schema.GetInstancesRequest{ServiceType: expectedServiceType}
		getInstancesResponse, err := e.client.GetInstances(context.Background(), getInstancesRequest)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		// there should be the only instance right now
		assert.Len(t, getInstancesResponse.ServiceInstances, 1)
		assert.Equal(t, expectedServiceInstance, getServicesResponse.ServiceTypes[0])

		// 2.1 validate http
		data, err = sendHTTPReq(e.serverCfg, "instances")
		require.NoError(t, err)

		httpGetInstancesResponse := &schema.GetInstancesResponse{}
		err = jsonpb.Unmarshal(data, httpGetInstancesResponse)
		require.NoError(t, err)
		require.True(t, proto.Equal(httpGetInstancesResponse, getInstancesResponse))

		// 3. ask for list of sessions
		getSessionsRequest := &schema.GetSessionsRequest{
			ServiceDescription: &schema.ServiceDescription{
				ServiceType:     expectedServiceType,
				ServiceInstance: expectedServiceInstance,
			},
		}
		getSessionResponse, err := e.client.GetSessions(context.Background(), getSessionsRequest)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		// there should be the only session right now
		assert.Len(t, getSessionResponse.Sessions, 1)
		session := getSessionResponse.Sessions[0]
		assert.Equal(t, expectedServiceType, session.Description.ServiceType)
		assert.Equal(t, expectedServiceInstance, session.Description.ServiceInstance)
		assert.Equal(t, uint32(0), session.Description.SessionId)

		// session start time must be greater than test start time
		sessionStartTime, err := ptypes.Timestamp(session.Metadata.StartedAt)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		assert.True(t, testStartTime.Before(sessionStartTime))

		// session is still operational, so the finish time is nil
		assert.Nil(t, session.Metadata.FinishedAt)

		// 3.1 validate http
		data, err = sendHTTPReq(e.serverCfg, "sessions")
		require.NoError(t, err)

		httpGetSessionResponse := &schema.GetSessionsResponse{}
		err = jsonpb.Unmarshal(data, httpGetSessionResponse)
		require.NoError(t, err)
		require.True(t, proto.Equal(httpGetSessionResponse, getSessionResponse))

		// 4. subscribe for session updates
		subscriptionRequest := &schema.SubscribeForSessionRequest{
			SessionDescription: session.Description,
		}
		subscription, err := e.client.SubscribeForSession(context.Background(), subscriptionRequest)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
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

func sendHTTPReq(cfg *config.Config, req string) (io.ReadCloser, error) {
	u, err := url.Parse("http://" + cfg.Frontend.FrontendEndpoint)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "v1", req)

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
