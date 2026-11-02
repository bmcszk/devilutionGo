package audio_test

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/audio"
)

// lazyOutcome bundles one lazy-load observation set.
type lazyOutcome struct {
	first, second       *audio.SfxSample
	firstErr, secondErr error
	source              *fakeSource
	catalog             *audio.SfxCatalog
	id                  audio.SfxID
}

// fakeSource counts reads so lazy loading can be asserted.
type fakeSource struct {
	mu    sync.Mutex
	reads []string
	err   error
}

func (f *fakeSource) ReadFile(name string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.err != nil {
		return nil, f.err
	}
	f.reads = append(f.reads, name)
	return []byte("RIFF"), nil
}

func (f *fakeSource) readCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.reads)
}

func TestNewSfxCatalogValidatesRows(t *testing.T) {
	for _, tc := range []struct {
		name    string
		effects []audio.Effect
		want    string
	}{
		{"empty id", []audio.Effect{{ID: "", Path: "sfx\\misc\\walk1.wav"}}, "empty id"},
		{"empty path", []audio.Effect{{ID: "Walk", Path: ""}}, "empty path"},
		{"duplicate", []audio.Effect{{ID: "Walk", Path: "a.wav"}, {ID: "Walk", Path: "b.wav"}}, "duplicate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// when the broken effect set is indexed
			catalog, err := audio.NewSfxCatalog(tc.effects, &fakeSource{})

			// then construction is refused with the defect named
			expectSfxCatalogRefused(t, catalog, err, tc.want)
		})
	}
}

func TestSfxCatalogPathsAreNonEmpty(t *testing.T) {
	// given a catalog of every id the sound code uses in the town scenario
	effects := []audio.Effect{
		{ID: string(audio.SfxWalk), Flags: []string{"Misc"}, Path: "sfx\\misc\\walk1.wav"},
		{ID: string(audio.SfxChestOpen), Flags: []string{"Misc"}, Path: "sfx\\misc\\chest.wav"},
	}
	catalog, err := audio.NewSfxCatalog(effects, &fakeSource{})
	if err != nil {
		t.Fatalf("NewSfxCatalog: %v", err)
	}

	// when every id is resolved
	paths := resolveSfxPaths(t, catalog)

	// then every id maps to a non-empty path and an unknown id is rejected
	if len(paths) != catalog.Len() {
		t.Fatalf("resolved %d ids, catalog holds %d", len(paths), catalog.Len())
	}
	expectNonEmptyPaths(t, paths)
	if _, err := catalog.Path("NoSuchEffect"); !errors.Is(err, audio.ErrUnknownSfx) {
		t.Fatalf("Path(NoSuchEffect) error = %v, want ErrUnknownSfx", err)
	}
}

func TestSfxCatalogLazyLoad(t *testing.T) {
	// given a catalog with one effect and a counting source
	source := &fakeSource{}
	catalog, err := audio.NewSfxCatalog([]audio.Effect{
		{ID: string(audio.SfxWalk), Path: "sfx\\misc\\walk1.wav"},
	}, source)
	if err != nil {
		t.Fatalf("NewSfxCatalog: %v", err)
	}
	id := audio.SfxID(audio.SfxWalk)

	// when the effect is loaded twice
	first, firstErr := catalog.Load(id)
	second, secondErr := catalog.Load(id)

	// then the source was read once, with an archive-style path, at 22050 Hz
	expectLazyLoad(t, lazyOutcome{
		first: first, second: second,
		firstErr: firstErr, secondErr: secondErr,
		source: source, catalog: catalog, id: id,
	})
}

func TestSfxCatalogCommittedEffectsTable(t *testing.T) {
	// given the committed base and Hellfire effect corpora
	base := mustParseSfxEffects(t, nil, "base")
	hf := mustParseSfxEffects(t, nil, "hf")

	// when a catalog is built from each
	baseCatalog := mustNewSfxCatalog(t, base)
	hfCatalog := mustNewSfxCatalog(t, hf)

	// then every row of both tables maps an id to a non-empty sfx path
	assertSfxPaths(t, "base", baseCatalog, 814)
	assertSfxPaths(t, "hf", hfCatalog, 917)
}

// mustParseSfxEffects parses one committed effects table.
func mustParseSfxEffects(t *testing.T, _ []byte, name string) []audio.Effect {
	t.Helper()

	rel := "assets/txtdata/sound/effects.tsv"
	if name == "hf" {
		rel = "mods/hf/txtdata/sound/effects.tsv"
	}
	effects, err := audio.ParseEffects(corpus(t, rel))
	if err != nil {
		t.Fatalf("ParseEffects(%s): %v", name, err)
	}
	return effects
}

// mustNewSfxCatalog indexes effects with a source that never reads.
func mustNewSfxCatalog(t *testing.T, effects []audio.Effect) *audio.SfxCatalog {
	t.Helper()

	catalog, err := audio.NewSfxCatalog(effects, audio.SfxSourceFunc(func(_ string) ([]byte, error) {
		return nil, errors.New("no assets in this test")
	}))
	if err != nil {
		t.Fatalf("NewSfxCatalog: %v", err)
	}
	return catalog
}

// assertSfxPaths asserts the catalog size and that every path is an sfx asset.
func assertSfxPaths(t *testing.T, name string, catalog *audio.SfxCatalog, want int) {
	t.Helper()

	if got := catalog.Len(); got != want {
		t.Fatalf("%s catalog size = %d, want %d", name, got, want)
	}
	for _, effect := range catalog.Effects() {
		if effect.Path == "" {
			t.Fatalf("%s effect %q has an empty path", name, effect.ID)
		}
		if !strings.HasPrefix(audio.ArchivePath(effect.Path), "sfx/") {
			t.Fatalf("%s effect %q path = %q, want an sfx/ asset", name, effect.ID, effect.Path)
		}
	}
}

// expectSfxCatalogRefused asserts construction reported want.
func expectSfxCatalogRefused(t *testing.T, catalog *audio.SfxCatalog, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("NewSfxCatalog accepted %+v", catalog)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want it to mention %q", err, want)
	}
}

// resolveSfxPaths resolves every catalog row to its path.
func resolveSfxPaths(t *testing.T, catalog *audio.SfxCatalog) map[audio.SfxID]string {
	t.Helper()

	paths := make(map[audio.SfxID]string, catalog.Len())
	for _, effect := range catalog.Effects() {
		path, err := catalog.Path(audio.SfxID(effect.ID))
		if err != nil {
			t.Fatalf("Path(%q): %v", effect.ID, err)
		}
		paths[audio.SfxID(effect.ID)] = path
	}
	return paths
}

// expectNonEmptyPaths asserts no id resolved to an empty path.
func expectNonEmptyPaths(t *testing.T, paths map[audio.SfxID]string) {
	t.Helper()

	for id, path := range paths {
		if path == "" {
			t.Fatalf("id %q maps to an empty path", id)
		}
	}
}

// expectLazyLoad asserts the sample was fetched once and described correctly.
func expectLazyLoad(t *testing.T, outcome lazyOutcome) {
	t.Helper()

	if outcome.firstErr != nil || outcome.secondErr != nil {
		t.Fatalf("Load errors = (%v, %v)", outcome.firstErr, outcome.secondErr)
	}
	if outcome.first != outcome.second {
		t.Fatal("second Load returned a different sample")
	}
	if got := outcome.source.readCount(); got != 1 {
		t.Fatalf("source reads = %d, want 1", got)
	}
	if want := "sfx/misc/walk1.wav"; strings.Join(outcome.source.reads, ",") != want {
		t.Fatalf("source read %v, want %q", outcome.source.reads, want)
	}
	if outcome.first.SampleRate != audio.DefaultSampleRate {
		t.Fatalf("sample rate = %d, want %d", outcome.first.SampleRate, audio.DefaultSampleRate)
	}
	if !outcome.catalog.Loaded(outcome.id) {
		t.Fatal("Loaded(id) = false after a successful load")
	}
}
