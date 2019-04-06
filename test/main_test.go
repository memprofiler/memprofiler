package test

import (
	"context"
	"testing"
	"time"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
	l, err := newLauncher()
	if err != nil {
		t.Fatal(err)
	}
	defer l.Stop()

	// wait until client will send couple of reports to server
	time.Sleep(2 * l.reporterCfg.Scenario.Steps[0].Wait.Duration)

	expectedServiceType := l.reporterCfg.Client.ServiceDescription.ServiceType
	expectedServiceInstance := l.reporterCfg.Client.ServiceDescription.ServiceInstance

	// 1. ask for list of service types
	getServicesRequest := &schema.GetServicesRequest{}
	getServicesResponse, err := l.client.GetServices(context.Background(), getServicesRequest)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	// there should be the only service right now
	assert.Len(t, getServicesResponse.ServiceTypes, 1)
	assert.Equal(t, expectedServiceType, getServicesResponse.ServiceTypes[0])

	// 2. ask for list of service instances
	getInstancesRequest := &schema.GetInstancesRequest{ServiceType: expectedServiceType}
	getInstancesResponse, err := l.client.GetInstances(context.Background(), getInstancesRequest)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Len(t, getInstancesResponse.ServiceInstances, 1)
	assert.Equal(t, expectedServiceInstance, getServicesResponse.ServiceTypes[0])
}
