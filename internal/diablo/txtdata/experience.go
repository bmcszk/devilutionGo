package txtdata

import (
	"fmt"
	"strconv"

	enginetxtdata "github.com/bmcszk/devilutionGo/internal/engine/txtdata"
)

// experienceThresholds is the level → XP threshold vector.
type experienceThresholds struct {
	thresholds []uint32
}

// loadExperienceThresholds parses Experience.tsv (header "Level,Experience").
func loadExperienceThresholds(t *enginetxtdata.Table) (experienceThresholds, error) {
	var e experienceThresholds
	li, err := t.Column("Level")
	if err != nil {
		return e, err
	}
	xi, err := t.Column("Experience")
	if err != nil {
		return e, err
	}
	for _, row := range t.Rows {
		level, xp, err := parseExpRow(row, li, xi)
		if err != nil {
			return e, err
		}
		// Levels missing from the column are filled with the row's own XP.
		for len(e.thresholds) < level {
			e.thresholds = append(e.thresholds, uint32(xp))
		}
		e.thresholds[level-1] = uint32(xp)
	}
	return e, nil
}

// parseExpRow parses one experience table row.
// Mirrors the per-row parsing in loadExperienceThresholds (experience.cpp).
func parseExpRow(row []string, li, xi int) (level int, xp uint64, err error) {
	if li >= len(row) || xi >= len(row) {
		return 0, 0, fmt.Errorf("txtdata: experience row too short: %v", row)
	}
	level, err = strconv.Atoi(row[li])
	if err != nil {
		return 0, 0, fmt.Errorf("txtdata: level %q: %w", row[li], err)
	}
	xp, err = strconv.ParseUint(row[xi], 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("txtdata: experience %q: %w", row[xi], err)
	}
	return level, xp, nil
}

// MaxLevel is the number of thresholds, capped at 255.
func (e experienceThresholds) MaxLevel() int {
	if len(e.thresholds) > 255 {
		return 255
	}
	return len(e.thresholds)
}

// ThresholdForLevel: level 0 → 0; levels past the table clamp to the max threshold.
func (e experienceThresholds) ThresholdForLevel(level int) uint32 {
	if level <= 0 {
		return 0
	}
	idx := level - 1
	if idx >= len(e.thresholds) {
		idx = len(e.thresholds) - 1
	}
	return e.thresholds[idx]
}
