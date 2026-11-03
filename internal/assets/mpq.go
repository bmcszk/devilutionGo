package assets

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// MPQ priorities are searched from the highest number down, so a larger value
// overrides a smaller one (assets.hpp:307).
const (
	// MainMpqPriority holds diabdat.mpq or spawn.mpq (assets.hpp:309).
	MainMpqPriority = 1000

	// HellfireMpqPriority holds hellfire.mpq (assets.cpp:581).
	HellfireMpqPriority = 8000

	// HFMonkMpqPriority holds the Hellfire monk graphics add-on (assets.cpp:591).
	HFMonkMpqPriority = 8100

	// HFBardMpqPriority holds the Hellfire bard add-on (assets.cpp:573).
	HFBardMpqPriority = 8110

	// HFBarbMpqPriority holds the Hellfire barbarian add-on (assets.cpp:574).
	HFBarbMpqPriority = 8120

	// HFMusicMpqPriority holds the Hellfire music add-on (assets.cpp:592).
	HFMusicMpqPriority = 8200

	// HFVoiceMpqPriority holds the Hellfire voice add-on (assets.cpp:593).
	HFVoiceMpqPriority = 8500

	// DevilutionXMpqPriority holds devilutionx.mpq, the port's own data (assets.hpp:310).
	DevilutionXMpqPriority = 9000

	// LangMpqPriority holds the selected translation archive (assets.hpp:311).
	LangMpqPriority = 9100

	// FontMpqPriority holds the selected font archive (assets.hpp:312).
	FontMpqPriority = 9200

	// ModMpqPriorityFirst is the priority of the first active mod; each further
	// mod takes the next integer (assets.cpp:768).
	ModMpqPriorityFirst = 10000
)

// mpqSource is one registered asset source: a packed MPQ archive or an unpacked
// directory tree standing in for one (assets.hpp:301-305 picks the underlying
// MpqArchiveT at build time; this stack keeps both behind the same interface).
type mpqSource struct {
	priority    int
	description string
	archive     *MPQArchive
	// read returns the asset bytes, or false when this source lacks them.
	read func(string) ([]byte, bool)
	// list returns the inner names this source holds, in its native spelling.
	list func() []string
}

// MPQStack resolves assets from MPQ archives and unpacked directory trees,
// searched from the highest priority down (assets.cpp:72-82).
//
// Safe for concurrent use: Registration guards byPriority/order with an
// RWMutex so readers (ReadFile/Files) hold only the read lock.
type MPQStack struct {
	// mu guards byPriority and order against concurrent registration and
	// lookup (brainer F8).
	mu sync.RWMutex
	// byPriority mirrors upstream's MpqArchives map, so a second registration
	// at an occupied priority keeps the first archive (assets.cpp:429-433).
	byPriority map[int]*mpqSource
	// order holds the registered sources by descending priority. It is rebuilt
	// only when a source is registered, so lookups allocate nothing.
	order []*mpqSource
}

// NewMPQStack returns an empty stack; register archives with Load or
// LoadUnpacked (MpqArchives starts empty, assets.cpp:43).
func NewMPQStack() *MPQStack {
	return &MPQStack{byPriority: map[int]*mpqSource{}}
}

// Load probes paths in order for "<name><ext>" and registers the first archive
// that opens at priority (assets.cpp:411-437; default ext ".mpq", uppercase
// ".MPQ" for the retail DIABDAT naming, assets.cpp:546). An archive that
// exists but fails to open is reported and the probe continues; false means
// no search path holds a readable archive.
func (s *MPQStack) Load(paths []string, name string, priority int, ext ...string) bool {
	suffix := ".mpq"
	if len(ext) == 1 {
		suffix = ext[0]
	}

	for _, dir := range paths {
		full := filepath.Join(dir, name+suffix)

		source, err := openMPQSource(full, priority)
		if err != nil {
			slog.Warn("mpq archive failed to open", "path", full, "error", err)

			continue
		}

		s.register(source)

		return true
	}

	slog.Warn("missing mpq archive", "name", name)

	return false
}

// LoadUnpacked registers the unpacked extraction "<path>/<name>/" as a source
// at priority, the UNPACKED_MPQS mode of Load (assets.cpp:392-396).
func (s *MPQStack) LoadUnpacked(paths []string, name string, priority int) bool {
	for _, dir := range paths {
		full := filepath.Join(dir, name) + string(os.PathSeparator)

		info, err := os.Stat(full)
		if err != nil || !info.IsDir() {
			continue
		}

		s.register(&mpqSource{
			priority:    priority,
			description: full,
			read:        dirReader(full),
			list:        dirLister(full),
		})

		return true
	}

	slog.Warn("missing unpacked mpq directory", "name", name)

	return false
}

// ReadFile returns the bytes of the asset name from the highest-priority source
// that holds it (assets.cpp:72-82). It satisfies AssetSource.
func (s *MPQStack) ReadFile(name string) ([]byte, bool) {
	asset := assetName(name)
	if asset == "" {
		return nil, false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, source := range s.order {
		if data, ok := source.read(asset); ok {
			return data, true
		}
	}

	return nil, false
}

// Files returns every inner asset name the stack serves, deduplicated
// case-insensitively and sorted. Names are slash-normalized (forward slashes);
// ReadFile accepts either separator. When two sources hold the same name, the
// higher-priority source names the entry, since the search order is descending
// by priority (assets.hpp:307).
func (s *MPQStack) Files() []string {
	s.mu.RLock()
	length := len(s.order)
	order := make([]*mpqSource, length)
	copy(order, s.order)
	s.mu.RUnlock()

	names := make([]string, 0, length)
	seen := make(map[string]struct{}, length)

	for _, source := range order {
		names = appendUnseen(names, seen, source.list())
	}

	slices.Sort(names)

	return names
}

// Close closes every open archive handle. Unpacked directories hold no handle
// and are skipped (assets.cpp:605-611 drops archives instead).
func (s *MPQStack) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	errs := make([]error, 0, len(s.order))

	for _, source := range s.order {
		if err := source.close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// close releases the archive handle a packed source holds; unpacked directory
// sources hold none.
func (src *mpqSource) close() error {
	if src.archive == nil {
		return nil
	}

	if err := src.archive.Close(); err != nil {
		return fmt.Errorf("close mpq %s: %w", src.description, err)
	}

	return nil
}

// register keeps the first source registered at a priority; a collision warns
// and keeps the existing one, like a failed emplace into MpqArchives
// (assets.cpp:429-433). The discarded archive is closed, as upstream's failed
// emplace destroys it.
func (s *MPQStack) register(source *mpqSource) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.byPriority == nil {
		s.byPriority = map[int]*mpqSource{}
	}

	if existing, occupied := s.byPriority[source.priority]; occupied {
		slog.Warn("mpq priority already registered, keeping the existing archive",
			"priority", source.priority, "skipped", source.description, "kept", existing.description)

		if err := source.close(); err != nil {
			slog.Warn("discarded mpq archive failed to close", "error", err)
		}

		return
	}

	s.byPriority[source.priority] = source
	s.rebuildOrder()
}

// rebuildOrder refreshes the descending-priority lookup order
// (assets.hpp:307 iterates MpqArchives, a std::greater<> keyed map).
func (s *MPQStack) rebuildOrder() {
	s.order = make([]*mpqSource, 0, len(s.byPriority))
	for _, source := range s.byPriority {
		s.order = append(s.order, source)
	}

	slices.SortFunc(s.order, func(a, b *mpqSource) int { return b.priority - a.priority })
}

// appendUnseen appends the slash-normalized names of listing that seen does not
// already hold, keyed case-insensitively so the first (highest-priority)
// spelling of a name wins.
func appendUnseen(names []string, seen map[string]struct{}, listing []string) []string {
	for _, entry := range listing {
		name := slashName(entry)
		key := strings.ToLower(name)

		if key == "" {
			continue
		}

		if _, duplicate := seen[key]; duplicate {
			continue
		}

		seen[key] = struct{}{}

		names = append(names, name)
	}

	return names
}

// slashName normalizes an archive inner path to the forward-slash spelling
// Files reports.
func slashName(name string) string {
	return strings.ReplaceAll(name, `\`, "/")
}

// assetName rejects names no lookup can resolve and converts an MPQ inner path
// ("dir\file.ext") to the host separator.
func assetName(name string) string {
	if name == "" || strings.HasSuffix(name, `\`) || strings.HasSuffix(name, "/") {
		return ""
	}

	return filepath.FromSlash(slashName(name))
}

// openMPQSource opens path and wraps it as a source at priority.
func openMPQSource(path string, priority int) (*mpqSource, error) {
	archive, err := OpenMPQArchive(path)
	if err != nil {
		return nil, fmt.Errorf("open mpq: %w", err)
	}

	return &mpqSource{
		priority:    priority,
		description: path,
		archive:     archive,
		read:        archiveReader(archive),
		list:        archiveLister(archive),
	}, nil
}

// archiveReader reads archive members through the record reader, which routes
// each sector's compression flags (IMPLODE included) through its decoder.
func archiveReader(archive *MPQArchive) func(string) ([]byte, bool) {
	return func(name string) ([]byte, bool) {
		data, err := archive.ReadFile(name)
		if err != nil {
			return nil, false
		}

		return data, true
	}
}

// archiveLister reads the archive's (listfile) in its stored spelling.
func archiveLister(archive *MPQArchive) func() []string {
	return func() []string {
		names, err := archive.Names()
		if err != nil {
			return nil
		}

		return names
	}
}

// dirReader reads a member from an unpacked extraction tree.
func dirReader(dir string) func(string) ([]byte, bool) {
	return func(name string) ([]byte, bool) {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, false
		}

		return data, true
	}
}

// dirLister walks an unpacked extraction tree for its member names, relative to
// dir.
func dirLister(dir string) func() []string {
	return func() []string {
		names, err := walkNames(dir)
		if err != nil {
			return nil
		}

		return names
	}
}

// walkNames collects the files under dir as slash-separated paths relative to
// dir.
func walkNames(dir string) ([]string, error) {
	var names []string

	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		relative, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return fmt.Errorf("relative path of %s: %w", path, relErr)
		}

		names = append(names, filepath.ToSlash(relative))

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", dir, err)
	}

	return names, nil
}
