package metrics

import (
	"context"

	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/common"
)

// Computer performs statistical analysis for the incoming and archived data streams
type Computer interface {
	// PutMeasurement registers new measurement for an actual session
	PutMeasurement(sd *schema.SessionDescription, mm *schema.Measurement) error
	// TODO: remove?
	SessionRecentMetrics(ctx context.Context, sd *schema.SessionDescription) (*schema.SessionMetrics, error)
	// SessionSubscribe returns new subscription for session updates
	SessionSubscribe(ctx context.Context, sd *schema.SessionDescription) (Subscription, error)
	// TODO: method to close session and free resources
	common.Subsystem
}

// Subscription provides push interface to receive actual session metrics
type Subscription interface {
	// Updates returns read-only channel with actual session metrics;
	// if the channel is closed, the session is terminated
	Updates() <-chan *schema.SessionMetrics
	// Unsubscribe frees resources occupied by subscription
	Unsubscribe()
	// publish sends message to the subscriber
	publish(*schema.SessionMetrics)
	// close terminates subscription
	close()
}

// dispatcher is a subscription manager
type dispatcher interface {
	// createSubscription creates new subscription for a session
	createSubscription(ctx context.Context, description *schema.SessionDescription) Subscription
	// dropSubscription deletes existing subscription
	dropSubscription(description *schema.SessionDescription, id subscriptionID)
	// broadcast sends message to all existing subscriptions
	broadcast(description *schema.SessionDescription, msg *schema.SessionMetrics)
}
