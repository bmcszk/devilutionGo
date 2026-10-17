// Command devilutionGo-headless is the headless entry point for CI and parity
// testing, built with the `headless` build tag.
//
// The -ticks flag runs the fixed-timestep game loop N ticks of 50ms each and
// exits 0, so CI can smoke-test the loop without a window or ebiten.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
)

func main() {
	ticks := flag.Int("ticks", 0, "run this many 50ms game ticks, then exit 0")
	flag.Parse()

	src := &engine.MockTickSource{}
	g := game.NewGame(src)
	for i := 0; i < *ticks; i++ {
		src.NowValue += 50 * time.Millisecond
		g.Update(50 * time.Millisecond)
	}
	_, _ = fmt.Println(*ticks)
	os.Exit(0)
}
