// Package assets loads game data from MPQ archives into a priority-keyed stack
// searched from highest priority down.
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

// Lookups walk from the highest priority down, so larger numbers override smaller ones.
const (
	FontMpqPriority        = 9200
	LangMpqPriority        = 9100
	DevilutionXMpqPriority = 9000
	MainMpqPriority        = 1000
)

// Stack resolves assets from archives ordered by priority.
//
// Not safe for concurrent use: archives are registered at startup and read on
// the main thread.
type Stack struct {
	archives map[int]*archive
}

type archive struct {
	priority    int
	description string
	mpqArchive  *mpq.Archive
	// reader returns the asset bytes, or false when this archive lacks it.
	reader func(string) ([]byte, bool)
}

// Open probes the search paths in order for "<name>.mpq" and registers the first
// readable hit at MainMpqPriority.
func Open(paths []string, name string) *Stack {
	stack := &Stack{archives: map[int]*archive{}}
	stack.Load(paths, name, MainMpqPriority)
	return stack
}

// Load probes the search paths in order for "<name>.mpq" and registers the first
// readable candidate at the given priority.
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

// LoadUnpacked registers "<path>/<name>/" as a directory archive at the given
// priority.
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

// ReadFile returns the asset bytes from the highest-priority archive that
// contains it.
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

func (s *Stack) sorted() []*archive {
	order := make([]*archive, 0, len(s.archives))
	for _, a := range s.archives {
		order = append(order, a)
	}
	slices.SortFunc(order, func(a, b *archive) int { return b.priority - a.priority })
	return order
}

// register keeps the first archive registered at a priority; a collision logs a
// warning and still reports success.
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

// normalizeAssetPath rejects empty names and trailing separators, and converts
// MPQ inner backslash paths to the OS separator.
func normalizeAssetPath(name string) string {
	if name == "" || strings.HasSuffix(name, `\`) || strings.HasSuffix(name, "/") {
		return ""
	}
	return filepath.FromSlash(strings.ReplaceAll(name, `\`, "/"))
}

func openMpq(path string, priority int) (*archive, error) {
	file, err := mpq.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	return &archive{priority: priority, description: path, mpqArchive: file, reader: mpqReader(file)}, nil
}

// mpqReader extracts to a temp file because go-mpq has no in-memory read.
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

func dirReader(dir string) func(string) ([]byte, bool) {
	return func(name string) ([]byte, bool) {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, false
		}
		return data, true
	}
}
