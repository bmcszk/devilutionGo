package audio

import "math"

// The sound-side positional math, ported 1:1 from the engine and kept pure
// stdlib: no device, no clock, no file io. The sim-purity gate scans this
// package, so everything the audio engine needs to DECIDE (volume, pan,
// audibility) lives here and the device half lives in cmd behind the windowed
// tag.

// Distance attenuation and pan bounds (Source/engine/sound_defs.hpp:9-16). They
// are duplicated from internal/engine, where the settings sliders need the same
// numbers: the two copies are pinned against each other by volume_test.go so a
// drift cannot pass.
const (
	// VolumeMin is the volume slider's floor: at it, sound and music switch off
	// (VOLUME_MIN, Source/engine/sound_defs.hpp:5).
	VolumeMin = -1600
	// VolumeMax is the volume slider's ceiling (VOLUME_MAX, :6).
	VolumeMax = 0
	// VolumeSteps is the Music/Sound slider granularity (VOLUME_STEPS, :7).
	VolumeSteps = 64
	// AttenuationMin is the distance attenuation floor: a sound whose computed
	// volume is at or below it is dropped (ATTENUATION_MIN, :9).
	AttenuationMin = -6400
	// AttenuationMax is the distance attenuation ceiling (:10).
	AttenuationMax = 0
	// PanMin and PanMax bound the stereo pan value (:13-14).
	PanMin = -6400
	PanMax = 6400
)

// VolumeLogToLinear constants (Source/utils/soundsample.cpp:46-60):
// VolumeScale is -1000 / log(0.5), the millibel span is -100..0 dB, and
// StereoSeparation is the 2:1 edge-mix tuning.
// They are float32 constants because upstream computes the whole curve in
// float (constexpr float VolumeScale = 3321.9281F, soundsample.cpp:46).
const (
	volumeScale      float32 = 3321.9281
	millibelMin      float32 = -10000
	millibelMax      float32 = 0
	stereoSeparation float32 = 6000
)

// Point is a tile position in the sound domain.
type Point struct {
	X int
	Y int
}

// CalculateSoundPosition converts a sound's tile into the log volume and pan it
// is heard at, from the player's tile (CalculateSoundPosition,
// Source/engine/sound_position.cpp:8-22).
//
// pan is always returned (upstream assigns it before the distance gate);
// audible is false when the sound falls outside the attenuation range, and the
// volume is then unusable.
func CalculateSoundPosition(playerPosition, soundPosition Point) (volume, pan int, audible bool) {
	delta := Point{X: soundPosition.X - playerPosition.X, Y: soundPosition.Y - playerPosition.Y}

	pan = clampInt((delta.X-delta.Y)*256, PanMin, PanMax)
	volume = ApproxDistance(playerPosition, soundPosition) * -64

	if volume <= AttenuationMin {
		return 0, pan, false
	}

	return volume, pan, true
}

// ApproxDistance is the integer tile distance with the upstream <5% error
// approximation (PointOf::ApproxDistance, Source/engine/point.hpp:118-130).
func ApproxDistance(from, to Point) int {
	dx := absInt(to.X - from.X)
	dy := absInt(to.Y - from.Y)

	minimum, maximum := dx, dy
	if minimum > maximum {
		minimum, maximum = maximum, minimum
	}

	approx := maximum*1007 + minimum*441
	if maximum < minimum*16 {
		approx -= maximum * 40
	}

	return (approx + 512) / 1024
}

// VolumeLogToLinear converts a log volume into a linear gain in [0,1]
// (VolumeLogToLinear, Source/utils/soundsample.cpp:69-74): the value is remapped
// onto the millibel span and then raised as 10^(mb/VolumeScale).
//
// Documented deviation: upstream evaluates std::pow over float, this port takes
// the double-precision pow and narrows, so the last float32 bit can differ. The
// golden fixture pins the port's own values, not a C++ oracle (no offline
// soundsample probe exists).
func VolumeLogToLinear(logVolume, logMin, logMax int) float32 {
	scaled := remap(float32(logMin), float32(logMax), millibelMin, millibelMax, float32(logVolume))

	// The division stays in float32, matching std::pow(LogBase, logScaled /
	// VolumeScale) where every operand is float.
	return float32(math.Pow(10, float64(scaled/volumeScale)))
}

// PanLogToLeftRight converts a log pan into per-channel gains
// (PanLogToLeftRight, Source/utils/soundsample.cpp:83-102): a centred sound is
// unity on both sides, a left sound attenuates the right channel, a right sound
// attenuates the left.
func PanLogToLeftRight(logPan int) (leftGain, rightGain float32) {
	if logPan == 0 {
		return 1, 1
	}

	// float32 division, as upstream (std::pow over float operands).
	factor := float32(math.Pow(10, float64(float32(-absInt(logPan))/stereoSeparation)))
	if logPan < 0 {
		return 1, factor
	}

	return factor, 1
}

// remap is math::Remap (Source/utils/math.h:66-70): outMin + (outMax-outMin) *
// InvLerp(inMin, inMax, v), with a zero-width input range mapping to outMin.
func remap(inMin, inMax, outMin, outMax, v float32) float32 {
	if inMax == inMin {
		return outMin
	}

	return outMin + (outMax-outMin)*((v-inMin)/(inMax-inMin))
}

// clampInt pins value into [low, high] (std::clamp).
func clampInt(value, low, high int) int {
	return min(max(value, low), high)
}

// absInt is std::abs for int.
func absInt(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
