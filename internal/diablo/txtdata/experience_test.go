package txtdata_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/txtdata"
)

// First rows copied from assets/txtdata/Experience.tsv (50 data rows total).
const expTSV = "Level\tExperience\r\n" +
	"1\t2000\r\n" +
	"2\t4620\r\n" +
	"3\t8040\r\n"

func TestLoadExperienceThresholdsGolden(t *testing.T) {
	// when
	e := mustLoadExperience(t)
	got := []uint32{e.ThresholdForLevel(1), e.ThresholdForLevel(2), e.ThresholdForLevel(3)}
	want := []uint32{2000, 4620, 8040}
	// then
	if !slices.Equal(got, want) {
		t.Fatalf("thresholds = %v, want %v", got, want)
	}
}

func TestMaxLevelMatchesRowCount(t *testing.T) {
	// then
	if got := mustLoadExperience(t).MaxLevel(); got != 50 {
		t.Fatalf("MaxLevel = %d, want 50 (V10.1: 50 rows)", got)
	}
}

func TestThresholdForLevelZeroAndClamp(t *testing.T) {
	e := mustLoadExperience(t)

	// when
	gotZero := e.ThresholdForLevel(0)
	gotPastMax := e.ThresholdForLevel(999)
	maxThreshold := e.ThresholdForLevel(50)

	// then
	if gotZero != 0 {
		t.Fatalf("level 0 threshold = %d, want 0", gotZero)
	}
	if gotPastMax != maxThreshold {
		t.Fatal("past-max level must clamp to max threshold")
	}
}

func TestLoadExperienceBadNumber(t *testing.T) {
	// when
	_, err := txtdata.LoadExperienceThresholds(mustParse(t, []byte("Level\tExperience\r\nX\tY\r\n")))

	// then
	if err == nil {
		t.Fatal("non-numeric experience = nil error, want error")
	}
}

func fullExperienceTSV() []byte {
	rows := []string{expTSV}
	for lvl := 4; lvl <= 50; lvl++ {
		rows = append(rows, strings.Join([]string{strconv.Itoa(lvl), strconv.Itoa(lvl * 1000)}, "\t")+"\r\n")
	}
	return []byte(strings.Join(rows, ""))
}

func mustLoadExperience(t *testing.T) txtdata.ExperienceThresholds {
	t.Helper()
	e, err := txtdata.LoadExperienceThresholds(mustParse(t, fullExperienceTSV()))
	if err != nil {
		t.Fatalf("txtdata.LoadExperienceThresholds: %v", err)
	}
	return e
}
