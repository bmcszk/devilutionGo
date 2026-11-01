// Package appdiag holds the Diablo-specific fatal-error messages that upstream
// raises through app_fatal (appfat.cpp) from game files. The engine appfat
// package stays free of game strings; this layer supplies them.
package appdiag

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/engine/appfat"
)

// InvalidEncodeParameters ports codec.cpp:144 app_fatal("Invalid encode
// parameters") — the save-game codec size guard.
func InvalidEncodeParameters() *appfat.FatalError {
	return appfat.AppFatal("Invalid encode parameters")
}

// WrongMissileID ports missiles.cpp:2920
// app_fatal(StrCat("wrong missile ID ", static_cast<int>(missile._mitype))).
func WrongMissileID(missileID int) *appfat.FatalError {
	return appfat.AppFatal(fmt.Sprintf("wrong missile ID %d", missileID))
}
