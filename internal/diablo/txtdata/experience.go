// Experience table per ReloadExperienceData
// (Source/tables/playerdat.cpp:89-156) and the ExperienceData helpers
// (playerdat.cpp:35-68, :403-405).
package txtdata

import (
	"fmt"
	"strconv"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// ExperienceThresholds is the level → XP threshold vector
// (playerdat.cpp:35-37 levelThresholds).
type ExperienceThresholds struct {
	thresholds []uint32
}

// LoadExperienceThresholds parses Experience.tsv (header "Level,Experience").
func LoadExperienceThresholds(t *enginetxtdata.Table) (ExperienceThresholds, error) {
	var e ExperienceThresholds
	li, err := t.Column("Level")
	if err != nil {
		return e, err
	}
	xi, err := t.Column("Experience")
	if err != nil {
		return e, err
	}
	for _, row := range t.Rows {
		if li >= len(row) || xi >= len(row) {
			return e, fmt.Errorf("txtdata: experience row too short: %v", row)
		}
		level, err := strconv.Atoi(row[li])
		if err != nil {
			return e, fmt.Errorf("txtdata: level %q: %w", row[li], err)
		}
		xp, err := strconv.ParseUint(row[xi], 10, 32)
		if err != nil {
			return e, fmt.Errorf("txtdata: experience %q: %w", row[xi], err)
		}
		// setThresholdForLevel indexes level-1 (playerdat.cpp:62-66); gaps in
		// the level column pad with the last threshold.
		for len(e.thresholds) < level {
			e.thresholds = append(e.thresholds, uint32(xp))
		}
		e.thresholds[level-1] = uint32(xp)
	}
	return e, nil
}

// MaxLevel mirrors getMaxLevel (playerdat.cpp:39-42): number of thresholds,
// capped at 255.
func (e ExperienceThresholds) MaxLevel() int {
	if len(e.thresholds) > 255 {
		return 255
	}
	return len(e.thresholds)
}

// ThresholdForLevel mirrors getThresholdForLevel (playerdat.cpp:48-55):
// level 0 → 0; levels past the table clamp to the max threshold.
func (e ExperienceThresholds) ThresholdForLevel(level int) uint32 {
	if level <= 0 {
		return 0
	}
	idx := level - 1
	if idx >= len(e.thresholds) {
		idx = len(e.thresholds) - 1
	}
	return e.thresholds[idx]
}
