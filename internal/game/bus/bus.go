// Package bus fans a finished kernel pass out to presentation consumers
// (renderer, sound, UI flourishes) over an in-process watermill GoChannel
// (ADR-018 law 5). Simulation-mutating handlers are called by the kernel
// itself; they never subscribe here, so the bus cannot perturb a pass.
package bus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"

	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// TopicEvents is the one topic the kernel publishes on: one message per pass.
const TopicEvents = "kernel.events"

// Publisher is the kernel.Publisher the presentation side subscribes to.
//
// Backend-doc caveat (watermill v1.5.3, pubsub/gochannel): Publish returns
// before delivery unless BlockPublishUntilSubscriberAck is set, every
// subscriber is fed from its own goroutine, and each gets a message copy - so
// cross-subscriber and cross-pass delivery order is not guaranteed. The kernel
// trace stays the authoritative order; one pass's batch keeps its own order.
type Publisher struct {
	channel *gochannel.GoChannel
}

// New returns a publisher on a fresh non-persistent GoChannel backend: a pass
// published while nobody listens is dropped, which is what presentation wants.
func New() *Publisher {
	return &Publisher{channel: gochannel.NewGoChannel(gochannel.Config{}, nil)}
}

// Publish sends one ordered pass as one message so the batch keeps the kernel's
// order. It is called by the kernel after the pass, never inside it.
//
// A publish failure is dropped on purpose: the bus is presentation only, and a
// dropped frame of display traffic must never fail a simulation pass.
func (p *Publisher) Publish(events []kernel.Event) {
	payload, err := json.Marshal(events)
	if err != nil {
		return // kernel.Event is integer-only, so this is unreachable.
	}

	_ = p.channel.Publish(TopicEvents, message.NewMessage(watermill.NewUUID(), payload))
}

// Close releases the subscribers; the kernel never calls it (it owns no bus).
func (p *Publisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("close bus: %w", err)
	}

	return nil
}

// Subscribe returns event batches, one per published pass, decoded into fresh
// copies: a consumer mutating what it receives cannot reach the kernel trace or
// another consumer. The channel closes when ctx is done or the bus closes.
//
// Delivery is asynchronous and per-subscriber (see Publisher), so consumers
// must treat a batch as "some finished pass", take its order from the batch,
// and take cross-pass order from the kernel trace, not from the channel.
func (p *Publisher) Subscribe(ctx context.Context) (<-chan []kernel.Event, error) {
	raw, err := p.channel.Subscribe(ctx, TopicEvents)
	if err != nil {
		return nil, fmt.Errorf("subscribe %q: %w", TopicEvents, err)
	}

	batches := make(chan []kernel.Event)

	go forward(ctx, raw, batches)

	return batches, nil
}

// forward decodes published messages into batches until ctx or the bus ends.
// A message is always acked: presentation traffic is dropped, never retried, so
// one malformed batch cannot stall the subscriber's send loop.
func forward(ctx context.Context, raw <-chan *message.Message, batches chan<- []kernel.Event) {
	defer close(batches)

	for msg := range raw {
		msg.Ack()

		events, err := decode(msg.Payload)
		if err != nil {
			continue
		}

		select {
		case batches <- events:
		case <-ctx.Done():
			return
		}
	}
}

func decode(payload []byte) ([]kernel.Event, error) {
	var events []kernel.Event

	if err := json.Unmarshal(payload, &events); err != nil {
		return nil, fmt.Errorf("decode events: %w", err)
	}

	return events, nil
}
