package light_test

import (
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/light"
)

// registryCallers is the independent copy of the measured inventory: every
// caller shape with its AddLight radius and the upstream sites it is measured
// from. It is kept next to the test so a wrong or missing registry row fails
// here instead of silently creating a light with the wrong radius.
var registryCallers = []struct {
	name   string
	source light.SourceType
	radius uint8
	cites  []string
}{
	{"player torch", light.SrcPlayerTorch, 10, []string{"player.cpp:2524"}},
	{"monster default", light.SrcMonsterDefault, 3, []string{"monster.cpp:3351"}},
	{"monster berserk", light.SrcMonsterBerserk, 3, []string{"missiles.cpp:1339"}},
	{"monster data light", light.SrcMonsterData, 3, []string{"diablo.cpp:1506"}},
	{"diablo", light.SrcDiablo, 8, []string{"monster.cpp:898"}},
	{"candle", light.SrcObjectCandle, 5, []string{"objects.cpp:1483"}},
	{"story candle", light.SrcObjectStoryCandle, 3, []string{"objects.cpp:4212"}},
	{"torch", light.SrcObjectTorch, 8, []string{"objects.cpp:4241"}},
	{"flame trap", light.SrcObjectFlameTrap, 1, []string{"objects.cpp:1582"}},
	{"rune", light.SrcMissileRune, 8, []string{"missiles.cpp:691"}},
	{"big explosion", light.SrcMissileBigExplosion, 8, []string{"missiles.cpp:1613", "missiles.cpp:1631"}},
	{"bolt", light.SrcMissileBolt, 5, []string{"missiles.cpp:1726", "missiles.cpp:1767", "missiles.cpp:2699"}},
	{"firebolt", light.SrcMissileFirebolt, 8, []string{"missiles.cpp:1876", "missiles.cpp:1906", "missiles.cpp:1987"}},
	{"lightning", light.SrcMissileLightning, 4, []string{"missiles.cpp:2015"}},
	{"faint", light.SrcMissileFaint, 1, []string{"missiles.cpp:2216", "missiles.cpp:2667"}},
	{"town portal", light.SrcTownPortal, 15, []string{"portal.cpp:59"}},
	{"firewall", light.SrcFireWall, 2, []string{"missiles.cpp:3055"}},
	{"flame wave", light.SrcFlameWave, 2, []string{"missiles.cpp:3470"}},
	{"weapon explosion", light.SrcWeaponExplosion, 9, []string{"missiles.cpp:3599"}},
	{"missile explosion light", light.SrcMissileExplosion, 9, []string{"missiles.cpp:3623"}},
}

// TestRegistryCoversEveryCaller pins the inventory size: every measured caller
// shape is registered, and the bean's "at least 12 distinct callers" floor
// holds.
func TestRegistryCoversEveryCaller(t *testing.T) {
	// given: the registry and the measured caller table.
	types := light.LightSourceTypes()

	// when
	got := len(types)

	// then
	if got < 12 {
		t.Fatalf("registry has %d sources, want at least 12", got)
	}
	if got != len(registryCallers) {
		t.Fatalf("registry has %d sources, want %d measured caller shapes", got, len(registryCallers))
	}
}

// TestRegistryRadii pins each row's radius against the measured AddLight call.
func TestRegistryRadii(t *testing.T) {
	// when
	got := make([]uint8, len(registryCallers))
	for i, caller := range registryCallers {
		got[i] = light.GetLightSourceDef(caller.source).Radius
	}

	// then
	for i, caller := range registryCallers {
		t.Run(caller.name, func(t *testing.T) {
			expectRadius(t, got[i], caller.radius)
		})
	}
}

// TestRegistryCitesEveryCaller pins that every row cites each upstream site it
// covers, so a row cannot drift from upstream un-noticed.
func TestRegistryCitesEveryCaller(t *testing.T) {
	// when
	got := make([]string, len(registryCallers))
	for i, caller := range registryCallers {
		got[i] = light.GetLightSourceDef(caller.source).Cite
	}

	// then
	for i, caller := range registryCallers {
		t.Run(caller.name, func(t *testing.T) {
			expectCites(t, got[i], caller.cites)
		})
	}
}

// TestRegistryAnimatedSources pins FlagAnimated against the three sources whose
// radius is re-set from an animation frame.
func TestRegistryAnimatedSources(t *testing.T) {
	// given: the animated sources (missiles.cpp:3055, :3470, objects.cpp:1596).
	sources := []light.SourceType{light.SrcFireWall, light.SrcFlameWave, light.SrcObjectFlameTrap}

	// when
	got := flagsOf(sources)

	// then
	for i, source := range sources {
		t.Run(sourceName(source), func(t *testing.T) {
			expectFlagSet(t, got[i], light.FlagAnimated)
		})
	}
}

// TestRegistryPersistentSources pins FlagPersistent on the lights that live as
// long as their owner (a player torch, a candle, a portal, a monster).
func TestRegistryPersistentSources(t *testing.T) {
	// given
	sources := []light.SourceType{light.SrcPlayerTorch, light.SrcObjectCandle, light.SrcTownPortal}

	// when
	got := flagsOf(sources)

	// then
	for i, source := range sources {
		t.Run(sourceName(source), func(t *testing.T) {
			expectFlagSet(t, got[i], light.FlagPersistent)
			expectFlagClear(t, got[i], light.FlagAnimated)
		})
	}
}

// TestRegistryNetworkedSources pins FlagNetworked on the monster lights, which
// the multiplayer delta re-places (msg.cpp:871, sync.cpp:199).
func TestRegistryNetworkedSources(t *testing.T) {
	// given
	sources := []light.SourceType{light.SrcMonsterDefault, light.SrcMonsterBerserk, light.SrcMonsterData}

	// when
	got := flagsOf(sources)

	// then
	for i, source := range sources {
		t.Run(sourceName(source), func(t *testing.T) {
			expectFlagSet(t, got[i], light.FlagNetworked)
		})
	}
}

// TestRegistryShortLivedMissileHasNoFlags pins a plain missile light: no
// animation ramp, no owner lifetime, no network mirror (missiles.cpp:1726).
func TestRegistryShortLivedMissileHasNoFlags(t *testing.T) {
	// given
	def := light.GetLightSourceDef(light.SrcMissileBolt)

	// when
	all := light.FlagAnimated | light.FlagPersistent | light.FlagNetworked

	// then
	if def.Flags&all != 0 {
		t.Fatalf("flags = %d, want no flag set", def.Flags)
	}
}

// TestGetLightSourceDefUnknown pins the miss path.
func TestGetLightSourceDefUnknown(t *testing.T) {
	// given: a type outside the registry.
	unknown := light.SourceType(len(light.LightSourceTypes()) + 100)

	// when
	def := light.GetLightSourceDef(unknown)

	// then
	if def != (light.LightSourceDef{}) {
		t.Fatalf("def = %+v, want the zero row", def)
	}
}

// TestAddSourceLightUsesRegistry pins the runtime bridge: the created light
// carries the registry radius for its source type.
func TestAddSourceLightUsesRegistry(t *testing.T) {
	// given
	list := light.NewLightList()

	// when
	id := list.AddSourceLight(light.SrcTownPortal, light.Point{X: 3, Y: 4})

	// then
	if id == light.NoLight {
		t.Fatal("AddSourceLight returned NoLight on an empty pool")
	}
	if got := list.Lights[id].Radius; got != 15 {
		t.Fatalf("radius = %d, want 15", got)
	}
}

// expectRadius fails when the registry radius for source differs from want.
func expectRadius(t *testing.T, got, want uint8) {
	t.Helper()

	if got != want {
		t.Fatalf("radius = %d, want %d", got, want)
	}
}

// expectCites fails when the row's citation misses one of the wanted sites.
func expectCites(t *testing.T, cite string, want []string) {
	t.Helper()

	for _, site := range want {
		if !strings.Contains(cite, site) {
			t.Fatalf("cite = %q, want it to contain %q", cite, site)
		}
	}
}

// flagsOf reads the registry flags for each source, in row order.
func flagsOf(sources []light.SourceType) []light.SourceFlag {
	flags := make([]light.SourceFlag, len(sources))
	for i, source := range sources {
		flags[i] = light.GetLightSourceDef(source).Flags
	}

	return flags
}

// sourceName returns a stable sub-test name for a source type.
func sourceName(source light.SourceType) string {
	for _, caller := range registryCallers {
		if caller.source == source {
			return caller.name
		}
	}

	return "unregistered"
}

// expectFlagSet fails when def does not carry flag.
func expectFlagSet(t *testing.T, flags, flag light.SourceFlag) {
	t.Helper()

	if flags&flag == 0 {
		t.Fatalf("flags = %d, want flag %d set", flags, flag)
	}
}

// expectFlagClear fails when def carries flag.
func expectFlagClear(t *testing.T, flags, flag light.SourceFlag) {
	t.Helper()

	if flags&flag != 0 {
		t.Fatalf("flags = %d, want flag %d clear", flags, flag)
	}
}
