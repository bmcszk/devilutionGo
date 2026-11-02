package audio

import (
	"errors"
	"fmt"
	"strings"
)

// DefaultSampleRate is the output rate devilutionX opens the audio device with:
// options.cpp:55 DEFAULT_AUDIO_SAMPLE_RATE, offered as {22050, 44100, 48000} at
// options.cpp:520. The engine audio layer passes it to the mixer; the catalog
// only records it so a loaded sample can be described.
const DefaultSampleRate = 22050

// sfxPathSeparator is the separator effects.tsv uses for its asset paths
// ("sfx\\misc\\walk1.wav"), because the corpus is authored for the Windows
// build (effects.cpp:154 `reader.readString("path", item.pszName)`).
const sfxPathSeparator = `\`

// ErrUnknownSfx is returned when an id is not present in the catalog.
var ErrUnknownSfx = errors.New("audio: unknown sfx id")

// SfxSource resolves an effect path to its raw .wav bytes. The MPQ loader backs
// it in production and tests back it with a fake; the catalog never touches the
// filesystem itself.
type SfxSource interface {
	ReadFile(name string) ([]byte, error)
}

// SfxSourceFunc adapts a function to SfxSource.
type SfxSourceFunc func(name string) ([]byte, error)

// SfxSample is one effect after its bytes have been fetched. Decoding to PCM is
// the engine audio layer's job (ebiten audio at DefaultSampleRate); the catalog
// hands over the container bytes and the path they came from.
type SfxSample struct {
	ID         SfxID
	Path       string
	Flags      []string
	Data       []byte
	SampleRate int
}

// SfxCatalog is the effects.tsv effect set keyed by id, with lazy, cached loads.
//
// Construction validates the corpus instead of trusting it: an empty id, an
// empty path, or a duplicate id means the table is broken and the catalog
// refuses to exist, so every later lookup can assume a non-empty path.
type SfxCatalog struct {
	effects []Effect
	byID    map[SfxID]Effect
	cache   map[SfxID]*SfxSample
	source  SfxSource
}

// ReadFile implements SfxSource.
func (f SfxSourceFunc) ReadFile(name string) ([]byte, error) { return f(name) }

// NewSfxCatalog indexes effects and validates every row.
func NewSfxCatalog(effects []Effect, source SfxSource) (*SfxCatalog, error) {
	catalog := &SfxCatalog{
		effects: make([]Effect, len(effects)),
		byID:    make(map[SfxID]Effect, len(effects)),
		cache:   make(map[SfxID]*SfxSample, len(effects)),
		source:  source,
	}
	copy(catalog.effects, effects)

	for _, effect := range catalog.effects {
		if err := catalog.index(effect); err != nil {
			return nil, err
		}
	}
	return catalog, nil
}

// Len reports how many effects the catalog holds.
func (c *SfxCatalog) Len() int {
	if c == nil {
		return 0
	}
	return len(c.effects)
}

// Effects returns the catalog rows in effects.tsv order.
func (c *SfxCatalog) Effects() []Effect {
	if c == nil {
		return nil
	}
	return c.effects
}

// Effect returns the row of one id.
func (c *SfxCatalog) Effect(id SfxID) (Effect, bool) {
	if c == nil {
		return Effect{}, false
	}
	effect, ok := c.byID[id]
	return effect, ok
}

// Path returns the declared asset path of one id. Paths are always non-empty
// because construction rejects empty ones.
func (c *SfxCatalog) Path(id SfxID) (string, error) {
	effect, ok := c.Effect(id)
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrUnknownSfx, id)
	}
	return effect.Path, nil
}

// ArchivePath converts the declared path to the forward-slash form an archive
// reader expects, preserving case ("sfx\misc\walk1.wav" → "sfx/misc/walk1.wav").
func ArchivePath(path string) string {
	return strings.ReplaceAll(path, sfxPathSeparator, "/")
}

// Loaded reports whether an id's bytes were already fetched.
func (c *SfxCatalog) Loaded(id SfxID) bool {
	if c == nil {
		return false
	}
	_, ok := c.cache[id]
	return ok
}

// Load fetches one effect's bytes, at most once per id (effects.cpp:180 keeps
// the loaded sound in sgSFX and only reloads when pSnd is null).
func (c *SfxCatalog) Load(id SfxID) (*SfxSample, error) {
	if sample, ok := c.cache[id]; ok && sample != nil {
		return sample, nil
	}

	path, err := c.Path(id)
	if err != nil {
		return nil, err
	}
	if c.source == nil {
		return nil, errors.New("audio: sfx catalog has no source")
	}

	data, err := c.source.ReadFile(ArchivePath(path))
	if err != nil {
		return nil, fmt.Errorf("audio: load sfx %q from %q: %w", id, path, err)
	}

	sample := &SfxSample{
		ID:         id,
		Path:       path,
		Flags:      c.byID[id].Flags,
		Data:       data,
		SampleRate: DefaultSampleRate,
	}
	c.cache[id] = sample
	return sample, nil
}

// index validates one row and records it.
func (c *SfxCatalog) index(effect Effect) error {
	id := SfxID(effect.ID)
	if id == "" {
		return errors.New("audio: effect row has an empty id")
	}
	if effect.Path == "" {
		return fmt.Errorf("audio: effect %q has an empty path", id)
	}
	if _, dup := c.byID[id]; dup {
		return fmt.Errorf("audio: duplicate effect id %q", id)
	}
	c.byID[id] = effect
	return nil
}
