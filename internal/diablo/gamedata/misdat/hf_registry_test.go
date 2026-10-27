package misdat_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"
)

// Empty names are upstream nullptr (no function), so only named functions count.
func TestHFRegistrySuperset(t *testing.T) {
	// given
	raw := hfCorpus(t)

	// when
	hf, err := misdat.ParseMisdat(raw)
	// then
	if err != nil {
		t.Fatalf("misdat.ParseMisdat hf: %v", err)
	}

	adds := namedSet(hf, func(m misdat.MissileData) string { return m.AddFn })
	procs := namedSet(hf, func(m misdat.MissileData) string { return m.ProcessFn })

	if len(adds) != 73 {
		t.Fatalf("hf distinct addFn = %d, want 73 (excl. empty)", len(adds))
	}

	if len(procs) != 45 {
		t.Fatalf("hf distinct processFn = %d, want 45 (excl. empty)", len(procs))
	}

	checkHFRegistered(t, adds)
	checkHFProcesses(t, procs)
}

// namedSet collects the distinct non-empty values picked by get.
func namedSet(rows []misdat.MissileData, get func(misdat.MissileData) string) map[string]bool {
	set := map[string]bool{}

	for _, m := range rows {
		if name := get(m); name != "" {
			set[name] = true
		}
	}

	return set
}

// checkHFRegistered fails on any hf addFn missing from both registries.
func checkHFRegistered(t *testing.T, adds map[string]bool) {
	t.Helper()

	for name := range adds {
		if _, ok := misdat.AddRegistry[name]; !ok {
			if _, ok := misdat.HFAddRegistry[name]; !ok {
				t.Errorf("hf addFn %q not registered", name)
			}
		}
	}
}

// checkHFProcesses fails on any hf processFn missing from the registry.
func checkHFProcesses(t *testing.T, procs map[string]bool) {
	t.Helper()

	for name := range procs {
		if _, ok := misdat.ProcessRegistry[name]; !ok {
			t.Errorf("hf processFn %q not registered", name)
		}
	}
}

func hfCorpus(t *testing.T) []byte {
	t.Helper()
	return corpus(t, "mods/hf/txtdata/missiles/misdat.tsv")
}
