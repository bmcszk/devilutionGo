// Package scene composes renderable game scenes at the presentation edge: it
// glues the Diablo level
// generators' grids (internal/diablo/levels) to the ported rasterizer
// (internal/render). The diablo layer may import both sides (ADR-008); the
// render layer stays level-package-free. First slice: the fixed town of
// devilGo-6asa, composed from the retail archives headlessly.
package scene
