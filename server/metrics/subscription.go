package metrics

import (
	"context"

	"github.com/memprofiler/memprofiler/schema"
)

type subscriptionID = uint64

var _ Subscription = (*defaultSubscription)(nil)

type defaultSubscription struct {
	sessionDescription *schema.SessionDescription  // helps to identify sessions
	id                 subscriptionID              // unique subscription id
	updates            chan *schema.SessionMetrics // channel to push data to client
	dispatcher         dispatcher                  // subscription dispatcher
	ctx                context.Context             // subscription context
}

func (s *defaultSubscription) Updates() <-chan *schema.SessionMetrics { return s.updates }

func (s *defaultSubscription) Unsubscribe() {
	s.dispatcher.dropSubscription(s.sessionDescription, s.id)
}

func (s *defaultSubscription) publish(msg *schema.SessionMetrics) {
	select {
	case s.updates <- msg:
	case <-s.ctx.Done():
	}
}

func (s *defaultSubscription) close() { close(s.updates) }

const updatesChanCapacity = 256

func newSubscription(
	ctx context.Context,
	id subscriptionID,
	sessionDescription *schema.SessionDescription,
	dispatcher dispatcher,
) Subscription {
	return &defaultSubscription{
		ctx:                ctx,
		sessionDescription: sessionDescription,
		id:                 id,
		dispatcher:         dispatcher,
		updates:            make(chan *schema.SessionMetrics, updatesChanCapacity),
	}
}
