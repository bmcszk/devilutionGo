// The AC-3 frame budget (devilGo-prcc): a frame-timing benchmark over the whole
// frame pipeline at the parity scene, and the p95 pin that turns the measured
// budget into an asserted one.
//
// Measurement bound (measured, not assumed): Presenter.Draw uploads its index and
// shade textures with Image.WritePixels, and outside a running ebiten loop the
// graphics command queue retains both buffers (2.4 MB per presented frame) until a
// loop flushes it. Frame times are stable for the first few hundred frames and
// drift up after that (measured: 4.3 ms/frame at frames 0-128, 15 ms/frame at
// frames 450+). Every measurement here is therefore bounded to frameSamples frames
// on a freshly built scene, and a long `-bench` run is a comparative tool, not
// budget evidence. Fixing the retention is a render-lane change, not this gate's.
package integration_test

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bmcszk/devilutionGo/internal/render"
)

const (
	// shortSamples is the sample count -short mode measures: 32 frames is enough
	// for a stable nearest-rank p95 without making the smoke slow.
	shortSamples = 32
	// benchFrameBudget is the p95 budget the pin asserts, in milliseconds: the AC-3
	// 60 fps frame time. A frame that misses it drops the rate below 60 on the box
	// that produced it.
	benchFrameBudget = budgetFrameMillis
	// ciBudget is the p95 budget on a shared CI runner, in milliseconds. CI runner
	// class measured 21.138ms p95 on run 44656997769; factor 1.6 ranks a real
	// regression (anything past 26.7ms) while tolerating the shared-runner class;
	// revisit on CI hardware changes.
	ciBudget = budgetFrameMillis * 1.6
)

// budgetClass is one runner class's p95 budget: the milliseconds the pin asserts
// and the class name a failure message reports.
type budgetClass struct {
	label  string
	millis float64
}

// BenchmarkFramePipeline times the full frame pipeline at the parity scene: the
// world pass, the panel stage, the text stage and the palette presenter run in
// draw order, one frame per iteration, with one sub-benchmark per stage plus the
// total. Each sub-benchmark builds its own scene, so its numbers are the scene's
// own and not the previous sub-benchmark's retained queue.
//
// The ns/op figures are comparative; the asserted budget is the p95 the pin
// measures (TestM1FrameBudgetP95), which the benchmark also logs per stage.
func BenchmarkFramePipeline(b *testing.B) {
	b.ReportAllocs()

	b.Run("total", func(b *testing.B) {
		scene := newFrameScene(b)
		// The p95 is measured first, on the fresh scene: it is the bounded sample
		// this gate pins, and b.Loop's own iteration count is the comparative ns/op.
		b.Logf("total p95 over %d frames: %.3fms", frameSamples, p95(measureFrames(scene, frameSamples).total))

		for b.Loop() {
			runFrame(scene, frameStages(), nil)
		}
	})

	for _, stage := range frameStages() {
		b.Run(stage.name, func(b *testing.B) {
			scene := newFrameScene(b)
			b.Logf("%s p95 over %d frames: %.3fms", stage.name, frameSamples, p95(measureStage(scene, stage)))

			for b.Loop() {
				stage.run(scene)
			}
		})
	}
}

// TestM1FrameBudgetP95 pins AC-3: the full frame pipeline's p95 frame time at the
// parity scene stays under the budget of the runner class it measures. The AC
// scopes the 16.67 ms 60 fps frame time to mid hardware, so the strict budget is
// the local baseline and a documented wider budget covers the shared CI runner
// class (frameBudgetClass).
//
// The series is sampled frameSamples times after frameWarmup warm-up frames, and
// the p95 is the nearest-rank value, so one scheduling spike cannot pass for a
// distribution. Short mode (-short) asserts the same budget over shortSamples.
func TestM1FrameBudgetP95(t *testing.T) {
	// given the parity scene, its frame pipeline and the budget of this runner class
	scene := newFrameScene(t)
	budget := frameBudgetClass()

	samples := frameSamples
	if testing.Short() {
		samples = shortSamples
	}

	// when the whole pipeline is sampled frame by frame
	times := measureFrames(scene, samples)

	// then the total p95 holds that budget
	if got := p95(times.total); got >= budget.millis {
		t.Fatalf("frame pipeline p95 = %.3fms over %d frames, budget %.3fms (%s): %s",
			got, samples, budget.millis, budget.label, stageTable(times))
	}

	// and every stage reports its own p95 under the same budget
	for name, series := range times.stages {
		if got := p95(series); got >= budget.millis {
			t.Fatalf("%s stage p95 = %.3fms over %d frames, budget %.3fms (%s): %s",
				name, got, samples, budget.millis, budget.label, stageTable(times))
		}
	}

	t.Logf("p95 over %d frames (budget %.3fms, %s): %s", samples, budget.millis, budget.label,
		stageTable(times))
}

// TestM1FramePipelineStagesAreNotEmpty pins that the pinned frame pipeline draws
// something: the world pass paints the layout, the panel stage paints its sheet,
// the text stage paints glyphs, and the surface handed to the palette pass carries
// a real spread of palette indices. A budget measured over a no-op pipeline would
// pass while proving nothing.
//
// The presented frame's colour readback needs a live ebiten loop (ReadPixels
// panics outside one), so the assertion stops at the index surface the presenter
// consumes.
func TestM1FramePipelineStagesAreNotEmpty(t *testing.T) {
	// given the parity scene and its pipeline
	scene := newFrameScene(t)
	stages := frameStages()

	// when every stage runs once in draw order
	clear(scene.surface.Pixels)

	painted := make(map[string]int, len(stages))
	before := make([]uint8, len(scene.surface.Pixels))

	for _, stage := range stages {
		copy(before, scene.surface.Pixels)
		stage.run(scene)

		painted[stage.name] = countChanged(before, scene.surface.Pixels)
	}

	// then the world, panel and text stages each changed the frame surface
	for _, name := range []string{stageWorld, stagePanel, stageText} {
		if painted[name] == 0 {
			t.Fatalf("%s stage changed 0 pixels of the frame, want a painted stage: %v", name, painted)
		}
	}

	// and the surface the palette pass resolves is not a flat fill
	if indices := distinctIndices(scene.surface); indices < 16 {
		t.Fatalf("frame surface carries %d distinct palette indices, want the layout's spread", indices)
	}

	t.Logf("pixels changed per stage: %v, distinct indices: %d", painted, distinctIndices(scene.surface))
}

// frameBudgetClass picks the budget of the runner class this process runs on: the
// declared mid-hardware baseline everywhere, the widened class where GitHub CI's
// own environment marks the runner. Both classes assert their own budget; neither
// branch skips the measurement.
func frameBudgetClass() budgetClass {
	if os.Getenv("CI") == "true" {
		return budgetClass{label: "CI runner", millis: ciBudget}
	}

	return budgetClass{label: "mid hardware, 60 fps", millis: benchFrameBudget}
}

// countChanged counts the positions where two equal-length buffers differ.
func countChanged(before, after []uint8) int {
	changed := 0
	for i := range before {
		if before[i] != after[i] {
			changed++
		}
	}

	return changed
}

// distinctIndices counts the distinct palette indices a surface carries.
func distinctIndices(surface render.Surface) int {
	seen := make(map[uint8]struct{}, 256)
	for _, index := range surface.Pixels {
		seen[index] = struct{}{}
	}

	return len(seen)
}

// measureStage samples one stage alone over a frame the pipeline just drew: the
// per-stage number a sub-benchmark reports, with the same warm-up and sample count
// the pipeline pin uses.
func measureStage(scene *frameScene, stage frameStage) []time.Duration {
	runFrame(scene, frameStages(), nil)

	series := make([]time.Duration, 0, frameSamples)
	for range frameSamples {
		begin := time.Now()
		stage.run(scene)
		series = append(series, time.Since(begin))
	}

	return series
}

// stageTable renders the per-stage plus total p95 table.
func stageTable(times *frameTimes) string {
	table := make([]string, 0, len(frameStages())+1)
	for _, stage := range frameStages() {
		table = append(table, stage.name+"="+formatMillis(p95(times.stages[stage.name])))
	}

	table = append(table, "total="+formatMillis(p95(times.total)))

	return strings.Join(table, " ")
}

// formatMillis renders one millisecond figure with three decimals: the panel and
// text stages sit in the microseconds, so two decimals would report them as zero.
func formatMillis(value float64) string {
	return strconv.FormatFloat(value, 'f', 3, 64) + "ms"
}
