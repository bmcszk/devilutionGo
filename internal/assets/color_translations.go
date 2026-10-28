package assets

// colorTranslationPaths is the font TRN table in upstream order
// (text_render.cpp:87-111): 19 slots, three of them null (indices 4, 5, 7).
var colorTranslationPaths = [19]string{
	`fonts\goldui.trn`,
	`fonts\grayui.trn`,
	`fonts\golduis.trn`,
	`fonts\grayuis.trn`,
	"", // ColorDialogWhite
	"", // ColorDialogRed
	`fonts\yellow.trn`,
	"",
	`fonts\black.trn`,
	`fonts\white.trn`,
	`fonts\whitegold.trn`,
	`fonts\red.trn`,
	`fonts\blue.trn`,
	`fonts\orange.trn`,
	`fonts\buttonface.trn`,
	`fonts\buttonpushed.trn`,
	`fonts\gamedialogwhite.trn`,
	`fonts\gamedialogyellow.trn`,
	`fonts\gamedialogred.trn`,
}

// ColorTranslations returns the font TRN paths by slot; an empty string marks a
// slot upstream leaves null (text_render.cpp:88-110).
func ColorTranslations() []string {
	paths := make([]string, len(colorTranslationPaths))
	copy(paths, colorTranslationPaths[:])

	return paths
}

// ColorTranslationPath returns the TRN path for a font colour slot; ok is false
// for the null slots (text_render.cpp:88-110).
func ColorTranslationPath(i int) (path string, ok bool) {
	if i < 0 || i >= len(colorTranslationPaths) || colorTranslationPaths[i] == "" {
		return "", false
	}

	return colorTranslationPaths[i], true
}

// ColorTranslationTRN loads a font colour's TRN through the optional-load path
// (trn.cpp-style LoadOptionalFileInMem); a null slot reports no TRN without an
// error, matching the lazy per-colour cache's "no table" state
// (text_render.cpp:113,190-192).
func (*TRNRegistry) ColorTranslationTRN(src AssetSource, i int) (TRN, bool, error) {
	path, ok := ColorTranslationPath(i)
	if !ok {
		return TRN{}, false, nil
	}

	return loadOptionalTRN(src, path)
}
