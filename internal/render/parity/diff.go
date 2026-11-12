package parity

// Diff compares two dumps: the index gate is byte-exact, the colour gate
// compares palette-expanded RGB and honours the allow-list. Colour expansion
// never masks an index mismatch: the index-exact law is not relaxable.
func Diff(a, b FrameDump, options Options) Report {
	report := Report{IndexExact: true}

	diffIndexGate(&report, a, b)

	diffColourGate(&report, a, b, options)

	return report
}

// diffIndexGate scans the index bytes row-major and records the first mismatch.
func diffIndexGate(report *Report, a, b FrameDump) {
	for i := range a.Width * a.Height {
		if a.Index[i] == b.Index[i] {
			continue
		}

		x, y := i%a.Width, i/a.Width
		report.IndexExact = false
		report.IndexDiff = &PixelDiff{
			X: x, Y: y,
			TileX: x / TileSize, TileY: y / TileSize,
			Got: b.Index[i], Want: a.Index[i],
		}

		return
	}
}

// diffColourGate compares the palette-expanded frames and records every
// unmasked mismatch in scan order.
func diffColourGate(report *Report, a, b FrameDump, options Options) {
	masked := make([]Region, 0, len(options.AllowList))

	for i := range a.Width * a.Height {
		got, want := b.expandAt(i), a.expandAt(i)
		if got == want {
			continue
		}

		x, y := i%a.Width, i/a.Width
		if region, maskedPixel := coveringRegion(options.AllowList, x, y); maskedPixel {
			masked = appendMaskedRegion(masked, region)

			continue
		}

		report.ColourDiffs = append(report.ColourDiffs, PixelDiff{
			X: x, Y: y,
			TileX: x / TileSize, TileY: y / TileSize,
			GotRGB: got, WantRGB: want,
		})
	}

	if len(masked) != 0 {
		report.Masked = masked
	}
}

// coveringRegion returns the first allow-list region covering the pixel.
func coveringRegion(regions []Region, x, y int) (Region, bool) {
	for _, region := range regions {
		if x >= region.X && x < region.X+region.Width && y >= region.Y && y < region.Y+region.Height {
			return region, true
		}
	}

	return Region{}, false
}

// appendMaskedRegion appends region unless an earlier entry already echoed it,
// so Masked holds distinct regions in input order.
func appendMaskedRegion(masked []Region, region Region) []Region {
	for _, seen := range masked {
		if seen == region {
			return masked
		}
	}

	return append(masked, region)
}
