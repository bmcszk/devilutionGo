package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	slashpath "path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bmcszk/devilutionGo/internal/assets"
	manifestpkg "github.com/bmcszk/devilutionGo/internal/assets/manifest"
)

// conversionInputs is the width oracle the extraction pass converts against
// (ADR-019): the geometry each legacy record needs, keyed by the CLX path that
// record converts to. An empty plan converts nothing.
type conversionInputs map[string]spriteGeometry

// spriteGeometry is what a CLX converter needs beyond the record bytes: the
// frame widths CEL and CL2 headers do not carry, and the frame count a PCX needs
// to split into frames.
type spriteGeometry struct {
	widths []int
	frames int
}

// conversionRecord is the manifest's record of the inputs a conversion consumed,
// so a reader can tell a converted file from an extracted one and rerun it.
type conversionRecord struct {
	Format string `yaml:"format"           json:"format"`
	Widths []int  `yaml:"widths,omitempty" json:"widths,omitempty"`
	Frames int    `yaml:"frames,omitempty" json:"frames,omitempty"`
	SHA256 string `yaml:"sha256"           json:"sha256"`
}

// conversionPlan keys the inventory geometry by the CLX path a record converts
// to. An entry carrying no frame widths cannot drive a conversion, so it gets no
// plan row.
func conversionPlan(inventory manifestpkg.Manifest) conversionInputs {
	plan := make(conversionInputs, len(inventory.Entries))

	for _, entry := range inventory.Entries {
		geometry, hasWidths := spriteGeometryOf(entry)
		if !hasWidths {
			continue
		}

		plan[conversionKey(entry.Path)] = geometry
	}

	return plan
}

// spriteGeometryOf reads the conversion geometry an inventory entry supplies: the
// per-frame widths when the entry lists them, else the single frame width.
func spriteGeometryOf(entry manifestpkg.Entry) (spriteGeometry, bool) {
	widths := entry.Widths
	if len(widths) == 0 && entry.Width > 0 {
		widths = []int{entry.Width}
	}

	if len(widths) == 0 {
		return spriteGeometry{}, false
	}

	return spriteGeometry{widths: widths, frames: entry.Frames}, true
}

// loadConversionInputs builds the conversion oracle from the graphics inventory.
// A machine without the inventory converts nothing and says so; an inventory that
// is present but unreadable fails the run, because converting on a guess would
// write widths nobody asserted.
func loadConversionInputs(path string) (conversionInputs, []string, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		notice := fmt.Sprintf("no inventory at %s: legacy records keep their source format", path)

		return nil, []string{notice}, nil
	}

	inventory, err := loadInventory(path)
	if err != nil {
		return nil, nil, err
	}

	return conversionPlan(inventory), nil, nil
}

// conversionKey normalizes a path for the plan lookup: inventory and archive
// names differ in case and separator, and a size-identical miss would silently
// skip a conversion.
func conversionKey(path string) string {
	return strings.ToLower(slashpath.Clean(strings.ReplaceAll(path, `\`, "/")))
}

// convertLegacyRecord replaces a legacy record's bytes with its CLX conversion
// when the plan describes the CLX path it would produce and a converter exists
// for its source format. Every other record keeps the copy-as-is payload, because
// nothing in the inventory says how wide its frames are (ADR-019).
func (inputs conversionInputs) convertLegacyRecord(name string, decoded rendered) (rendered, error) {
	sourceFormat := formatOf(strings.ToLower(filepath.Ext(name)))
	target := changesExtension(decoded.outputName, ".clx")

	geometry, planned := inputs[conversionKey(target)]
	if !planned || !validFormat(sourceFormat) {
		return decoded, nil
	}

	converted, err := geometry.convert(sourceFormat, decoded.payload)
	if err != nil {
		return rendered{}, fmt.Errorf("convert %s: %w", decoded.outputName, err)
	}

	list, err := assets.DecodeCLX(converted)
	if err != nil {
		return rendered{}, fmt.Errorf("decode converted %s: %w", decoded.outputName, err)
	}

	return rendered{
		payload:    converted,
		outputName: target,
		family:     familySprite,
		format:     "clx",
		frames:     clxLayout(list),
		conversion: geometry.conversionRecord(sourceFormat),
	}, nil
}

// convert turns one legacy record's bytes into CLX with the inventory geometry.
func (g spriteGeometry) convert(sourceFormat string, data []byte) ([]byte, error) {
	switch sourceFormat {
	case "cel":
		return convertCEL(data, g.widths)
	case "cl2":
		return convertCL2(data, g.widths)
	default:
		// A PCX carries no widths file: its frame count divides the image
		// height, and the inventory has no field for a transparent palette index.
		return convertPCX(data, g.frames, -1)
	}
}

// conversionRecord names the inputs a conversion consumed.
func (g spriteGeometry) conversionRecord(sourceFormat string) *conversionRecord {
	return &conversionRecord{
		Format: sourceFormat,
		Widths: g.widths,
		Frames: g.frames,
		SHA256: g.digest(sourceFormat),
	}
}

// digest hashes the conversion inputs, so the manifest identifies what was fed to
// the converter and not only its output.
func (g spriteGeometry) digest(sourceFormat string) string {
	widths := make([]string, len(g.widths))
	for i, width := range g.widths {
		widths[i] = strconv.Itoa(width)
	}

	canonical := fmt.Sprintf("%s\t%s\t%d", sourceFormat, strings.Join(widths, ","), g.frames)
	sum := sha256.Sum256([]byte(canonical))

	return hex.EncodeToString(sum[:])
}
