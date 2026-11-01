// Command devilutionGo is the GUI entry point for the devilutionGo engine.
package main

import (
	"log/slog"
	"os"

	"github.com/bmcszk/devilutionGo/internal/engine"
	"github.com/bmcszk/devilutionGo/internal/game"
)

func main() {
	// Upstream only ships the cheat commands in _DEBUG builds; here the same
	// surface is an explicit opt-in switch (--debug).
	g := game.NewGame(&engine.MockTickSource{})
	g.Debug = game.NewDebugRouterFromFlag(os.Args[1:], game.DebugEnv{})

	slog.Info("devilutionGo started", "debug", g.Debug != nil)
}
