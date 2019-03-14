package metrics

import (
	"context"
	"sync"

	"github.com/memprofiler/memprofiler/schema"
)

var _ dispatcher = (*defaultDispatcher)(nil)

// defaultDispatcher is a type used for subscription management
type defaultDispatcher struct {
	// sessionDescriptionID -> subscriptionID -> subscription
	subscriptions map[string]map[subscriptionID]Subscription
	// counter is a latest given subscriptionID
	counter subscriptionID
	mutex   sync.RWMutex
}

func (d *defaultDispatcher) createSubscription(
	ctx context.Context,
	sd *schema.SessionDescription) Subscription {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	sessionID := shortSessionIdentifier(sd)
	ss, exist := d.subscriptions[sessionID]
	if !exist {
		ss = make(map[subscriptionID]Subscription)
		d.subscriptions[sessionID] = ss
	}

	// create and store new subscription
	d.counter++
	subscription := newSubscription(ctx, d.counter, sd, d)
	ss[d.counter] = subscription
	return subscription
}

func (d *defaultDispatcher) dropSubscription(
	sd *schema.SessionDescription,
	id subscriptionID) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	sessionID := shortSessionIdentifier(sd)

	// close subscription and delete it from map
	d.subscriptions[sessionID][id].close()
	delete(d.subscriptions[sessionID], id)

	// clear top-level map if necessary
	if len(d.subscriptions[sessionID]) == 0 {
		delete(d.subscriptions, sessionID)
	}
}

func (d *defaultDispatcher) broadcast(sd *schema.SessionDescription, msg *schema.SessionMetrics) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	sessionID := shortSessionIdentifier(sd)
	ss, exist := d.subscriptions[sessionID]
	if !exist {
		return
	}

	// publish message to every subscriber
	for _, s := range ss {
		s.publish(msg)
	}
}

func newDispatcher() dispatcher {
	return &defaultDispatcher{
		subscriptions: make(map[string]map[subscriptionID]Subscription),
		mutex:         sync.RWMutex{},
	}
}
