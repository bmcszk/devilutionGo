package txtdata

import (
	"fmt"
	"strconv"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// ExperienceThresholds is the level → XP threshold vector.
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
		// Levels missing from the column are filled with the row's own XP.
		for len(e.thresholds) < level {
			e.thresholds = append(e.thresholds, uint32(xp))
		}
		e.thresholds[level-1] = uint32(xp)
	}
	return e, nil
}

// MaxLevel is the number of thresholds, capped at 255.
func (e ExperienceThresholds) MaxLevel() int {
	if len(e.thresholds) > 255 {
		return 255
	}
	return len(e.thresholds)
}

// ThresholdForLevel: level 0 → 0; levels past the table clamp to the max threshold.
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
