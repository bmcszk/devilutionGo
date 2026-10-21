// Command devilutionGo-headless is the headless entry point for CI and parity
// testing.
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
)

const tickDuration = 50 * time.Millisecond

func main() {
	ticks := flag.Int("ticks", 0, "run this many 50ms game ticks, then exit 0")
	flag.Parse()

	src := &engine.MockTickSource{}
	g := game.NewGame(src)
	for i := 0; i < *ticks; i++ {
		src.NowValue += tickDuration
		g.Update(tickDuration)
	}
	_, _ = fmt.Println(*ticks)
}
