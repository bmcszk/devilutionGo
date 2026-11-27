//go:build windowed

package main

import (
	"context"
	"log/slog"

	snd "github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audio"
	"github.com/bmcszk/devilutionGo/internal/game/bus"
	"github.com/bmcszk/devilutionGo/internal/game/kernel"
)

// audio_bus.go wires the FIRST kernel presentation subscriber (ADR-018 law 5):
// the bus fans a finished pass out, and the audio sink plays whatever sound the
// pass named. Nothing here can reach the sim: the kernel publishes AFTER the
// pass and a subscriber only reads a copy.

// sfxForEvent maps a kernel event onto the sound it names.
//
// Today NO kernel event carries a sound. The enumerated Kinds
// (internal/game/kernel/command.go:30-150) are sim decisions — stick push,
// attacks, loot, save/load, tick rate — with integer payloads only, and upstream
// plays their effects inside the sim handlers (Source/effects.cpp:208-232), not
// from an event. The table therefore maps nothing; a sound-bearing Kind becomes
// a one-line addition here, and the subscriber below is the wiring that will
// play it.
func sfxForEvent(ev kernel.Event) (snd.SfxID, bool) {
	_ = ev

	return "", false
}

// startAudioBus installs the publisher on the kernel and starts the audio
// subscriber. The returned cancel stops the subscriber goroutine.
func startAudioBus(
	ctx context.Context, engine *kernel.Kernel, sink *audioEngine,
) (*bus.Publisher, context.CancelFunc) {
	publisher := bus.New()
	engine.SetPublisher(publisher)

	subCtx, cancel := context.WithCancel(ctx)

	batches, err := publisher.Subscribe(subCtx)
	if err != nil {
		slog.Warn("windowed boot: audio bus subscribe", "err", err)
		cancel()

		return publisher, func() {}
	}

	go playBatchedSounds(subCtx, batches, sink)

	return publisher, cancel
}

// playBatchedSounds plays every sound-bearing event of every published pass.
func playBatchedSounds(ctx context.Context, batches <-chan []kernel.Event, sink *audioEngine) {
	for {
		select {
		case <-ctx.Done():
			return
		case events, ok := <-batches:
			if !ok {
				return
			}

			playEventSounds(events, sink)
		}
	}
}

// playEventSounds plays the events that name a sound, in batch order.
func playEventSounds(events []kernel.Event, sink *audioEngine) {
	for _, ev := range events {
		id, ok := sfxForEvent(ev)
		if !ok {
			continue
		}

		if err := sink.PlaySFX(id); err != nil {
			slog.Warn("windowed boot: bus sfx", "id", id, "err", err)
		}
	}
}
