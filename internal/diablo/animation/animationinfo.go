// Package animation ports AnimationInfo (Source/engine/animationinfo.h +
// animationinfo.cpp) and the nthread progress variable it reads.
package animation

// BaseValueFraction mirrors AnimationInfo::baseValueFraction: fractions are
// fixed point with 128 == 100%.
const BaseValueFraction = 128

// DistFlags mirrors AnimationDistributionFlags (animationinfo.h:17-30).
type DistFlags uint8

const (
	FlagNone DistFlags = 0
	// FlagProcessAnimationPending: processAnimation will be called after
	// setNewAnimation in the same game tick.
	FlagProcessAnimationPending DistFlags = 1 << 0
	// FlagSkipsDelayOfLastFrame: the delay of the last frame is ignored.
	FlagSkipsDelayOfLastFrame DistFlags = 1 << 1
	// FlagRepeatedAction: repeated animation (melee attack chains).
	FlagRepeatedAction DistFlags = 1 << 2
)

// ProgressSink carries the nthread global ProgressToNextGameTick
// (nthread.h:22): the fixed-point fraction of time to the next game tick.
// Upstream this is a mutable global; here the caller injects it so tests
// and the engine loop share one source of truth.
type ProgressSink struct {
	ProgressToNextGameTick uint8
}

// Info mirrors class AnimationInfo (animationinfo.h:36-118).
type Info struct {
	// NumberOfFrames: frames in the current animation.
	NumberOfFrames int8
	// CurrentFrame: current frame index into the sequence.
	CurrentFrame int8
	// TicksPerFrame: game ticks needed to advance one frame.
	TicksPerFrame int8
	// TickCounterOfCurrentFrame: counts up to TicksPerFrame.
	TickCounterOfCurrentFrame int8
	// IsPetrified: animation frozen (no interpolation).
	IsPetrified bool

	relevantFramesForDistributing int8
	skippedFramesFromPreviousAnim int8
	tickModifier                  uint16
	ticksSinceSequenceStarted     int16
}

// FrameToUseForRendering ports getFrameToUseForRendering
// (animationinfo.cpp:11-63).
func (a *Info) FrameToUseForRendering(g *ProgressSink) int {
	if a.relevantFramesForDistributing <= 0 {
		return maxInt8(int(a.CurrentFrame), 0)
	}
	if a.CurrentFrame >= a.relevantFramesForDistributing {
		return int(a.CurrentFrame)
	}

	ticksSinceStart := a.ticksSinceSequenceStarted
	if ticksSinceStart < 0 {
		ticksSinceStart = 0
	}

	totalTicks := int32(a.ProgressToNextGameTickOf(g)) + int32(ticksSinceStart)
	absolute := int(totalTicks * int32(a.tickModifier) / BaseValueFraction / BaseValueFraction)

	return a.computeFrameIndex(absolute)
}

// computeFrameIndex computes the final frame index from the absolute tick count.
// Mirrors the downstream logic of getFrameToUseForRendering (animationinfo.cpp:11-63).
func (a *Info) computeFrameIndex(absolute int) int {
	if a.skippedFramesFromPreviousAnim > 0 {
		absolute -= int(a.skippedFramesFromPreviousAnim)
		if absolute < 0 {
			absolute = int(a.NumberOfFrames) + absolute
		}
	} else if absolute >= int(a.relevantFramesForDistributing) {
		return int(a.relevantFramesForDistributing) - 1
	}
	if absolute < 0 {
		return 0
	}
	return absolute
}

// AnimationProgress ports getAnimationProgress (animationinfo.cpp:65-88).
func (a *Info) AnimationProgress(g *ProgressSink) uint8 {
	ticksSinceSequenceStarted := maxI16(a.ticksSinceSequenceStarted, 0)
	tickModifier := int32(a.tickModifier)
	if a.relevantFramesForDistributing <= 0 {
		if a.TicksPerFrame <= 0 {
			return 0
		}
		ticksSinceSequenceStarted = (int16(a.CurrentFrame)*int16(a.TicksPerFrame) +
			int16(a.TickCounterOfCurrentFrame)) * BaseValueFraction
		tickModifier = BaseValueFraction / int32(a.TicksPerFrame)
	}
	totalTicks := int32(a.ProgressToNextGameTickOf(g)) + int32(ticksSinceSequenceStarted)
	progressInFrames := totalTicks * tickModifier
	fraction := progressInFrames / int32(a.NumberOfFrames) / BaseValueFraction
	return uint8(fraction)
}

// NewAnimationOptions carries the optional arguments of
// setNewAnimation (animationinfo.h:95): everything after numberOfFrames and
// ticksPerFrame is optional upstream with defaults None/0/0/0.
type NewAnimationOptions struct {
	// Flags mirrors the AnimationDistributionFlags parameter
	// (animationinfo.h:95, default None).
	Flags DistFlags
	// NumSkippedFrames mirrors numSkippedFrames
	// (animationinfo.h:95, default 0).
	NumSkippedFrames int
	// DistributeFramesBeforeFrame mirrors distributeFramesBeforeFrame
	// (animationinfo.h:95, default 0).
	DistributeFramesBeforeFrame int
	// PreviewShownGameTickFragments mirrors previewShownGameTickFragments:
	// how long (in game tick fractions) the preview animation was shown
	// (animationinfo.h:93, default 0).
	PreviewShownGameTickFragments uint8
}

// SetNewAnimation ports setNewAnimation (animationinfo.cpp:90-190). The
// sprites argument is omitted: sprite binding is the render layer's job.
func (a *Info) SetNewAnimation(numberOfFrames, ticksPerFrame int, opts NewAnimationOptions) {
	flags := opts.Flags
	numSkippedFrames := opts.NumSkippedFrames
	distributeFramesBeforeFrame := opts.DistributeFramesBeforeFrame
	previewShownGameTickFragments := opts.PreviewShownGameTickFragments
	repeated := flags&FlagRepeatedAction == FlagRepeatedAction && distributeFramesBeforeFrame != 0 &&
		a.NumberOfFrames == int8(numberOfFrames) &&
		int(a.CurrentFrame)+1 >= distributeFramesBeforeFrame &&
		int(a.CurrentFrame) != int(a.NumberOfFrames)-1
	if repeated {
		a.skippedFramesFromPreviousAnim = a.NumberOfFrames - a.CurrentFrame - 1
	} else {
		a.skippedFramesFromPreviousAnim = 0
	}
	if ticksPerFrame <= 0 {
		ticksPerFrame = 1
	}
	a.NumberOfFrames = int8(numberOfFrames)
	a.CurrentFrame = int8(numSkippedFrames)
	a.TickCounterOfCurrentFrame = 0
	a.TicksPerFrame = int8(ticksPerFrame)
	a.ticksSinceSequenceStarted = 0
	a.relevantFramesForDistributing = 0
	a.tickModifier = 0
	a.IsPetrified = false

	if numSkippedFrames == 0 && flags == FlagNone {
		return
	}
	a.applyDistribution(numberOfFrames, ticksPerFrame, flags,
		numSkippedFrames, distributeFramesBeforeFrame, previewShownGameTickFragments)
}

// applyDistribution is the distribution branch of setNewAnimation
// (animationinfo.cpp:131-188), extracted verbatim.
func (a *Info) applyDistribution(
	numberOfFrames, ticksPerFrame int, flags DistFlags,
	numSkippedFrames, distributeFramesBeforeFrame int,
	previewShownGameTickFragments uint8,
) {
	relevantFrames := numberOfFrames
	if distributeFramesBeforeFrame != 0 {
		// Skipped frames are never distributed after the hit frame
		// (_pAFNum/_pSFNum): the weapon must point at the enemy on hit.
		relevantFrames = distributeFramesBeforeFrame - 1
	}
	relevantTicks := relevantFrames * ticksPerFrame
	relevantTicksSkipping := relevantTicks - numSkippedFrames*ticksPerFrame

	if flags&FlagProcessAnimationPending == FlagProcessAnimationPending {
		// processAnimation will run in this same game tick: pre-correct the
		// sequence counter so rendering never sees a negative tick offset.
		relevantTicksSkipping--
		a.ticksSinceSequenceStarted = -BaseValueFraction
	}
	if flags&FlagSkipsDelayOfLastFrame == FlagSkipsDelayOfLastFrame {
		// Game logic checks the frame, not the delay: the last frame would
		// otherwise show shorter than the others.
		relevantTicksSkipping -= ticksPerFrame - 1
	}
	relevantTicks += int(a.skippedFramesFromPreviousAnim) * ticksPerFrame

	relevantTicksFixed := int32(relevantTicks) * BaseValueFraction
	relevantTicksSkippingFixed := int32(relevantTicksSkipping) * BaseValueFraction

	a.ticksSinceSequenceStarted += int16(previewShownGameTickFragments)
	relevantTicksSkippingFixed += int32(previewShownGameTickFragments)

	tickModifier := int32(0)
	if relevantTicksSkippingFixed != 0 {
		tickModifier = BaseValueFraction * relevantTicksFixed / relevantTicksSkippingFixed
	}
	tickModifier /= int32(ticksPerFrame)

	a.relevantFramesForDistributing = int8(relevantFrames)
	a.tickModifier = uint16(tickModifier)
}

// ChangeAnimationData ports changeAnimationData
// (animationinfo.cpp:192-206): on-the-fly gear-change re-binding.
func (a *Info) ChangeAnimationData(numberOfFrames, ticksPerFrame int) {
	if numberOfFrames != int(a.NumberOfFrames) || ticksPerFrame != int(a.TicksPerFrame) {
		a.CurrentFrame = clampFrame(int(a.CurrentFrame), numberOfFrames)
		a.NumberOfFrames = int8(numberOfFrames)
		a.TicksPerFrame = int8(ticksPerFrame)
		a.ticksSinceSequenceStarted = 0
		a.relevantFramesForDistributing = 0
		a.tickModifier = 0
	}
}

// clampFrame mirrors the std::clamp / -1 branch in changeAnimationData.
func clampFrame(current, numberOfFrames int) int8 {
	if numberOfFrames < 1 {
		return -1
	}
	if current < 0 {
		return 0
	}
	if current > numberOfFrames-1 {
		return int8(numberOfFrames - 1)
	}
	return int8(current)
}

// ProcessAnimation ports processAnimation (animationinfo.cpp:208-228),
// forward playback.
func (a *Info) ProcessAnimation() {
	a.TickCounterOfCurrentFrame++
	a.ticksSinceSequenceStarted += BaseValueFraction
	if a.TickCounterOfCurrentFrame >= a.TicksPerFrame {
		a.TickCounterOfCurrentFrame = 0
		a.wrapFrameForward()
	}
}

// ProcessAnimationReverse is processAnimation with reverseAnimation=true
// (unseen-monster fading plays backwards).
func (a *Info) ProcessAnimationReverse() {
	a.TickCounterOfCurrentFrame++
	a.ticksSinceSequenceStarted += BaseValueFraction
	if a.TickCounterOfCurrentFrame >= a.TicksPerFrame {
		a.TickCounterOfCurrentFrame = 0
		a.wrapFrameReverse()
	}
}

// wrapFrameReverse wraps the frame index at the sequence start (reverse).
func (a *Info) wrapFrameReverse() {
	a.CurrentFrame--
	if a.CurrentFrame == -1 {
		a.CurrentFrame = a.NumberOfFrames - 1
		a.ticksSinceSequenceStarted = 0
	}
}

// wrapFrameForward wraps the frame index at the sequence end.
func (a *Info) wrapFrameForward() {
	a.CurrentFrame++
	if a.CurrentFrame >= a.NumberOfFrames {
		a.CurrentFrame = 0
		a.ticksSinceSequenceStarted = 0
	}
}

// ProgressToNextGameTickOf returns the interpolation fraction
// (getProgressToNextGameTick, animationinfo.cpp:224-229).
func (a *Info) ProgressToNextGameTickOf(g *ProgressSink) uint8 {
	if a.IsPetrified {
		return 0
	}
	return g.ProgressToNextGameTick
}

func maxInt8(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxI16(a, b int16) int16 {
	if a > b {
		return a
	}
	return b
}
