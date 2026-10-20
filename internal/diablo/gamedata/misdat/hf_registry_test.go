package misdat_test

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/misdat"

	"testing"
)

// hf build registers 73 named adds + 45 named processes (the bean's 74/46
// counts include the empty-name control rows, upstream nullptr).
func TestHFRegistrySuperset(t *testing.T) {
	hf, err := misdat.ParseMisdat(hfCorpus(t))
	if err != nil {
		t.Fatalf("misdat.ParseMisdat hf: %v", err)
	}
	adds := map[string]bool{}
	procs := map[string]bool{}
	for _, m := range hf {
		if m.AddFn != "" {
			adds[m.AddFn] = true
		}
		if m.ProcessFn != "" {
			procs[m.ProcessFn] = true
		}
	}
	if len(adds) != 73 {
		t.Fatalf("hf distinct addFn = %d, want 73 (excl. empty)", len(adds))
	}
	if len(procs) != 45 {
		t.Fatalf("hf distinct processFn = %d, want 45 (excl. empty)", len(procs))
	}
	for name := range adds {
		if _, ok := misdat.AddRegistry[name]; !ok {
			if _, ok := misdat.HFAddRegistry[name]; !ok {
				t.Errorf("hf addFn %q not registered", name)
			}
		}
	}
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
