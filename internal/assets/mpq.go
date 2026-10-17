// Package assets loads game data from MPQ archives, mirroring the C++ engine's
// assets.cpp: a priority-keyed stack searched from highest priority down.
package assets

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	mpq "github.com/suprsokr/go-mpq"
)

// Priority constants mirror assets.hpp:311-314. Lookups walk from the highest
// priority down, so larger numbers override smaller ones.
const (
	FontMpqPriority        = 9200
	LangMpqPriority        = 9100
	DevilutionXMpqPriority = 9000
	MainMpqPriority        = 1000
)

// Stack is the port of assets.hpp:307
// `std::map<int, MpqArchiveT, std::greater<>> MpqArchives`.
// Not safe for concurrent use; upstream registers archives at startup and reads on the main thread.
type Stack struct {
	archives map[int]*archive
}

// archive is one opened MPQ archive or unpacked directory registered at a priority.
type archive struct {
	priority    int
	description string
	mpqArchive  *mpq.Archive
	// reader returns the full bytes of the asset, or false when this archive
	// does not contain it. Directory archives capture their root in the closure.
	reader func(string) ([]byte, bool)
}

// Open probes search paths in order for "<name>.mpq" (assets.cpp:400-401) and
// registers the first readable hit at MainMpqPriority.
func Open(paths []string, name string) *Stack {
	stack := &Stack{archives: map[int]*archive{}}
	stack.Load(paths, name, MainMpqPriority)
	return stack
}

// Load mirrors assets.cpp:410-437 LoadMPQ: probe the search paths in order for
// "<name>.mpq", open the first readable candidate, and on a priority collision
// keep the existing registration.
func (s *Stack) Load(paths []string, name string, priority int) bool {
	for _, path := range paths {
		full := filepath.Join(path, name+".mpq")
		a, err := openMpq(full, priority)
		if err != nil {
			continue
		}
		return s.register(a)
	}
	log.Printf("Missing: %s", name)
	return false
}

// LoadUnpacked mirrors the UNPACKED_MPQS branch (assets.cpp:365-380
// FindUnpackedMpqData, 385-393 LoadMPQ): "<path>/<name>/" as a directory is
// registered verbatim and outranks packed archives at lower priorities.
func (s *Stack) LoadUnpacked(paths []string, name string, priority int) bool {
	for _, path := range paths {
		dir := filepath.Join(path, name) + string(os.PathSeparator)
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		return s.register(&archive{priority: priority, description: dir, reader: dirReader(dir)})
	}
	log.Printf("Missing unpacked MPQ directory: %s", name)
	return false
}

// ReadFile mirrors FindMpqFile (assets.cpp:78-89): walk the stack from highest
// priority down; the first archive containing the asset serves the bytes.
func (s *Stack) ReadFile(name string) ([]byte, bool) {
	asset := normalizeAssetPath(name)
	if asset == "" {
		return nil, false
	}
	for _, a := range s.sorted() {
		if data, ok := a.reader(asset); ok {
			return data, true
		}
	}
	return nil, false
}

// Close releases every open archive handle.
func (s *Stack) Close() {
	for _, a := range s.archives {
		if a.mpqArchive != nil {
			_ = a.mpqArchive.Close()
		}
	}
}

// sorted mirrors the std::greater<> map ordering: priorities descending.
func (s *Stack) sorted() []*archive {
	order := make([]*archive, 0, len(s.archives))
	for _, a := range s.archives {
		order = append(order, a)
	}
	slices.SortFunc(order, func(a, b *archive) int { return b.priority - a.priority })
	return order
}

// register mirrors assets.cpp:425-429: first registration at a priority wins;
// collisions log an error, keep the existing archive, and still return true —
// upstream LoadMPQ returns true because the search-path probe did find the MPQ.
func (s *Stack) register(a *archive) bool {
	if s.archives == nil {
		s.archives = map[int]*archive{}
	}
	if existing, occupied := s.archives[a.priority]; occupied {
		log.Printf("MPQ with priority %d is already registered, skipping %s", a.priority, existing.description)
		return true
	}
	s.archives[a.priority] = a
	return true
}

// normalizeAssetPath mirrors FindAsset (assets.cpp:192-194): reject empty names
// and trailing backslashes, and convert backslashes to the OS separator
// (upstream replaces '\\' with '/' on non-Windows; MPQ inner paths come in
// backslash form and unpacked trees are forward-slash on POSIX).
func normalizeAssetPath(name string) string {
	if name == "" || strings.HasSuffix(name, `\`) || strings.HasSuffix(name, "/") {
		return ""
	}
	return filepath.FromSlash(strings.ReplaceAll(name, `\`, "/"))
}

// openMpq opens "<path>/<name>.mpq" for reading; mechanism, no business rules.
func openMpq(path string, priority int) (*archive, error) {
	file, err := mpq.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	return &archive{priority: priority, description: path, mpqArchive: file, reader: mpqReader(file)}, nil
}

// mpqReader extracts an asset to a temp file and returns its bytes, because
// go-mpq exposes no in-memory read (gap recorded in the tracking bean).
func mpqReader(file *mpq.Archive) func(string) ([]byte, bool) {
	return func(name string) ([]byte, bool) {
		if !file.HasFile(name) {
			return nil, false
		}
		tmp, err := os.CreateTemp("", "dg-mpq-*.asset")
		if err != nil {
			return nil, false
		}
		tmpPath := tmp.Name()
		_ = tmp.Close()
		if err := file.ExtractFile(name, tmpPath); err != nil {
			_ = os.Remove(tmpPath)
			return nil, false
		}
		data, err := os.ReadFile(tmpPath)
		_ = os.Remove(tmpPath)
		if err != nil {
			return nil, false
		}
		return data, true
	}
}

// dirReader reads assets from an unpacked "<name>/" directory tree.
func dirReader(dir string) func(string) ([]byte, bool) {
	return func(name string) ([]byte, bool) {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, false
		}
		return data, true
	}
}
