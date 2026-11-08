package render_test

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/leveldat"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

const (
	// fullFade is the fade level of an undimmed frame (palette.cpp:140-147).
	fullFade = 256
	// parityFade is the dimmed fade level the parity scene presents at.
	parityFade = 200
	// paletteFixture is the committed Cathedral palette dump the tests load.
	paletteFixture = "l1_1.pal.hex"
	// captureDirPattern names the throwaway capture program's directory.
	captureDirPattern = "capture-"
	// captureDone marks the end of the draw-count capture program's output.
	captureDone = "CAPTURE-DONE"
	// captureFrames is how many frames the capture program renders before it
	// stops, so the logged frames are fully warmed up.
	captureFrames = "6"
	// drawBudget is the draw-command ceiling for one presentation frame.
	drawBudget = 40
	// bannedColorAPI is the colour-scale API the world index path must not reach
	// for (docs/research/dg-graphics.md G10).
	bannedColorAPI = "ColorScale ColorM ColorScaleScale"
	// captureMainSource is the throwaway program the draw-count test builds: one
	// full 640x480 presentation frame per Draw call, then termination.
	captureMainSource = `package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/diablo/gamedata/leveldat"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

type game struct {
	frame int
	surf  render.Surface
	pres  *render.Presenter
}

func (g *game) Update() error {
	g.frame++
	if g.frame < ` + captureFrames + ` {
		return nil
	}

	fmt.Printf("capture: %s\n", g.pres)

	return ebiten.Termination
}

func (g *game) Draw(screen *ebiten.Image) {
	g.pres.Draw(g.surf, nil, render.KeyTile)
	screen.DrawImage(g.pres.Target(), nil)
}

func (g *game) Layout(_, _ int) (int, int) { return 640, 480 }

func loadPalette() assets.Palette {
	body, err := os.ReadFile(filepath.Join("internal", "testdata", "render", "pal", "l1_1.pal.hex"))
	if err != nil {
		panic(err)
	}

	flat, err := hex.DecodeString(strings.Join(strings.Fields(string(body)), ""))
	if err != nil {
		panic(err)
	}

	pal, err := assets.LoadPaletteFromBytes(flat)
	if err != nil {
		panic(err)
	}

	return *pal
}

func main() {
	state := render.NewPaletteState(loadPalette(), light.MakeLightTables(leveldat.DTCathedral), assets.BlendingFor(0))

	surface := render.NewSurface(640, 480)
	for i := range surface.Pixels {
		surface.Pixels[i] = uint8(i % 251)
	}

	g := &game{surf: surface, pres: render.NewPresenter(state)}

	if err := ebiten.RunGame(g); err != nil && err != ebiten.Termination {
		fmt.Println("capture error:", err)
	}

	fmt.Println("` + captureDone + `")
}
`
)

var (
	// bannedColorNames is the banned-API list the G10 guard scans for.
	bannedColorNames = strings.Fields(bannedColorAPI)

	// customShaderDraw matches a draw whose shader entry carries no name: every
	// Ebitengine builtin logs as "shader: N (clear)", "shader: N (nearest)" and so
	// on, while a compiled Kage program logs its bare id, so this line is one
	// palette-lookup pass.
	customShaderDraw = regexp.MustCompile(`(?m)^\s+draw-triangles: .*\bshader: \d+\s*$`)
)

// loopGame runs one assertion inside the update callback of a live Ebitengine
// loop, where GPU readback is legal.
type loopGame struct {
	assert func() error
	frame  int
	done   chan error
}

// TestFinalPassDrawBudget captures the ebitenginedebug command log for one
// presentation frame in a child process and asserts the frame stays under the
// 40-command budget with exactly one custom-shader draw (the s2.log artifact).
func TestFinalPassDrawBudget(t *testing.T) {
	// given
	if !gpuAvailable(t) {
		t.Skipf("draw-count capture needs a GPU-capable run loop; %s CI has none (see s2.log notes)", runtime.GOOS)
	}

	// when
	log := captureDebugFrame(t)

	// then
	expectWithinDrawBudget(t, log)
}

// TestPaletteVariantsTextureMatchesDerivedVariants pins the upload contract: the
// bytes presented to the GPU as the 256x16 variant texture are the palette state
// itself, for every level type.
func TestPaletteVariantsTextureMatchesDerivedVariants(t *testing.T) {
	// given
	pal := loadPalette(t, paletteFixture)

	cases := []struct {
		name  string
		dtype leveldat.DungeonType
	}{
		{"town", leveldat.DTTown},
		{"cathedral", leveldat.DTCathedral},
		{"catacombs", leveldat.DTCatacombs},
		{"caves", leveldat.DTCaves},
		{"hell", leveldat.DTHell},
		{"nest", leveldat.DTNest},
		{"crypt", leveldat.DTCrypt},
	}

	// when
	textures := make([][]byte, len(cases))
	states := make([]render.PaletteState, len(cases))
	for i, tc := range cases {
		states[i] = render.NewPaletteState(pal, light.MakeLightTables(tc.dtype), assets.BlendingFor(int(tc.dtype)))
		textures[i] = render.PaletteVariantsTexture(&states[i].Variants, fullFade)
	}

	// then
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectVariantsTexture(t, textures[i], &states[i].Variants)
		})
	}
}

// TestPresentationPass pins the GPU palette-lookup pass against the tier-1
// software oracle: the fade arithmetic, the blended draw and the sprite-versus-
// tile keying trap, all read back from the render target inside one live run
// loop (scrollrt.cpp:252 ClxDrawLight, :267 ClxDrawLightBlended).
//
// Every GPU assertion shares one loop because a process hosts exactly one
// Ebitengine display session; a second RunGame panics.
func TestPresentationPass(t *testing.T) {
	// given
	if !gpuAvailable(t) {
		t.Skipf("presentation pass needs a GPU-capable run loop; %s CI has none (see s2.log notes)", runtime.GOOS)
	}

	// when
	err := inRunLoop(t, func() error {
		expectPresenterMatchesSoftwareOracle(t)
		expectKeyingTrap(t)

		return nil
	})
	// then
	if err != nil {
		t.Fatalf("presentation pass: %v", err)
	}
}

// TestWorldIndexPathHasNoColorScale guards the G10 law: the production index path
// must not reach for ebiten's colour scale or colour matrix, because the palette
// lookup happens in the shader.
func TestWorldIndexPathHasNoColorScale(t *testing.T) {
	// given
	entries, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob sources: %v", err)
	}

	// when
	production := productionSources(entries)

	// then
	for _, file := range production {
		expectNoColorAPI(t, file)
	}
}

// Update runs the assertion on the third frame and then stops the loop.
func (g *loopGame) Update() error {
	g.frame++
	if g.frame < 3 {
		return nil
	}

	err := g.assert()
	g.done <- err

	return ebiten.Termination
}

// Draw clears the screen: the presentation pass renders into its own target.
func (*loopGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{})
}

// Layout sizes the window to the world logical surface.
func (*loopGame) Layout(_, _ int) (width, height int) { return 640, 480 }

// gpuAvailable reports whether this machine can run the Ebitengine presentation
// pass: the shader compiles against GL and needs a live run loop. CI runners
// have no display, so the GPU-backed assertions degrade to a documented skip
// there (the compile-only and software-path assertions always run).
func gpuAvailable(t *testing.T) bool {
	t.Helper()

	display := os.Getenv("DISPLAY")
	if display == "" {
		return false
	}

	if _, err := os.Stat("/tmp/.X11-unix"); err != nil {
		return false
	}

	return true
}

// inRunLoop runs one Ebitengine loop and executes assert inside its update
// callback. A process can host exactly one loop, so every caller gates on
// gpuAvailable and the GPU assertions share TestPresentationPass.
func inRunLoop(t *testing.T, assert func() error) error {
	t.Helper()

	game := &loopGame{assert: assert, done: make(chan error, 1)}
	go func() {
		game.done <- ebiten.RunGame(game)
	}()

	if err := <-game.done; err != nil && !errors.Is(err, ebiten.Termination) {
		return fmt.Errorf("run loop: %w", err)
	}

	return nil
}

// expectPresenterMatchesSoftwareOracle compares the presented frame with the
// software oracle on a scene covering a fade level, a keyed draw and a blended
// draw. The fade lives in the uploaded variant texture, so both oracles expand
// through the undimmed variants and are dimmed the same way.
func expectPresenterMatchesSoftwareOracle(t *testing.T) {
	t.Helper()

	state, surface := oracleScene(t)

	tileOracle := surface.ExpandTile(&state.Variants, 0)
	applyFadeRGBA(tileOracle.Pixels, parityFade)

	spriteOracle := surface.ExpandSprite(&state.Variants, 0)
	applyFadeRGBA(spriteOracle.Pixels, parityFade)

	presenter := render.NewPresenter(state, render.WithFadeLevel(parityFade))

	expectFrameEqual(t, presenter.Frame(surface, nil, render.KeyTile), tileOracle)
	expectFrameEqual(t, presenter.Frame(surface, nil, render.KeySprite), spriteOracle)
}

// expectKeyingTrap asserts the G4 keying trap through the full pass: index-0
// pixels of a keyed draw are transparent while the same index in a tile draw
// writes the shade row's black.
func expectKeyingTrap(t *testing.T) {
	t.Helper()

	pal := loadPalette(t, paletteFixture)
	state := render.NewPaletteState(pal, light.MakeLightTables(leveldat.DTCathedral), assets.BlendingFor(0))

	surface := render.NewSurface(2, 1)
	surface.Set(0, 0, 0)
	surface.Set(1, 0, 1)

	presenter := render.NewPresenter(state)

	sprite := presenter.Frame(surface, nil, render.KeySprite)
	if sprite.Pixels[0] != (render.RGBA{}) {
		t.Errorf("sprite keying: pixel 0 = %v, want fully transparent", sprite.Pixels[0])
	}

	tileBlack := render.RGBA{R: pal[0][0], G: pal[0][1], B: pal[0][2], A: 0xFF}

	tile := presenter.Frame(surface, nil, render.KeyTile)
	if tile.Pixels[0] != tileBlack {
		t.Errorf("tile keep: pixel 0 = %v, want palette black %v", tile.Pixels[0], tileBlack)
	}
}

// productionSources drops the test files from a package listing: the guard
// covers the shipped index path, and this test names the banned API itself.
func productionSources(entries []string) []string {
	production := make([]string, 0, len(entries))
	for _, file := range entries {
		if !strings.HasSuffix(file, "_test.go") {
			production = append(production, file)
		}
	}

	return production
}

// expectNoColorAPI fails when one source file references the banned colour API.
func expectNoColorAPI(t *testing.T, file string) {
	t.Helper()

	body, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}

	for _, banned := range bannedColorNames {
		if bytes.Contains(body, []byte(banned)) {
			t.Errorf("%s references %s on the world index path, want none (G10)", file, banned)
		}
	}
}

// expectWithinDrawBudget asserts one presentation frame costs fewer draw commands
// than the budget and exactly one custom-shader draw.
func expectWithinDrawBudget(t *testing.T, log string) {
	t.Helper()

	draws := countDrawTriangles(log)
	if draws == 0 {
		t.Fatalf("ebitenginedebug log holds no draw-triangles commands; captured:\n%s", log)
	}

	if draws >= drawBudget {
		t.Fatalf("presentation frame used %d draw commands, want < %d\nlog:\n%s", draws, drawBudget, log)
	}

	if paletteDraws := countCustomShaderDraws(log); paletteDraws != 1 {
		t.Fatalf("final pass used %d custom-shader draws, want exactly 1\nlog:\n%s", paletteDraws, log)
	}
}

// oracleScene builds the blended test surface and its palette state.
func oracleScene(t *testing.T) (render.PaletteState, render.Surface) {
	t.Helper()

	pal := loadPalette(t, paletteFixture)
	tables := light.MakeLightTables(leveldat.DTCathedral)
	state := render.NewPaletteState(pal, tables, assets.BlendingFor(0))

	surface := render.NewSurface(32, 8)
	for y := range surface.Height {
		for x := range surface.Width {
			surface.Set(x, y, uint8(x*8+y+1))
		}
	}

	// sprite strip with keyed zeros, drawn at shade 3 (ClxDrawLight)
	surface.DrawDirect(clipWithLevels(1, 1, 4, 2, []byte{0, 5, 9, 0}, []byte{3, 3, 3, 3}))
	// blended strip at shade 12 (ClxDrawLightBlended)
	surface.DrawBlendedWithLightmap(
		clipWithLevels(1, 5, 4, 1, []byte{7, 7, 7, 7}, []byte{12, 12, 12, 12}), &tables.Tables, state.LUT)

	return state, surface
}

// captureDebugFrame writes the throwaway capture program into the module, runs
// it with -tags=ebitenginedebug against the live display and returns the
// Graphics commands section of its busiest presentation frame.
func captureDebugFrame(t *testing.T) string {
	t.Helper()

	root := moduleRoot(t)
	dir, err := os.MkdirTemp(root, captureDirPattern)
	if err != nil {
		t.Fatalf("create capture dir: %v", err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(captureMainSource), 0o600); err != nil {
		t.Fatalf("write capture main: %v", err)
	}

	cmd := exec.CommandContext(t.Context(), "go", "run", "-tags=ebitenginedebug", "./"+filepath.Base(dir))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "DISPLAY=:0")

	out, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(out), captureDone) {
		t.Fatalf("capture run: %v\n%s", err, out)
	}

	section := busiestFrame(string(out))
	if section == "" {
		t.Fatalf("capture output holds no populated Graphics commands section:\n%s", out)
	}

	return section
}

// moduleRoot returns the module directory of this test package.
func moduleRoot(t *testing.T) string {
	t.Helper()

	abs, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve module root: %v", err)
	}

	return abs
}

// busiestFrame returns the frame section carrying the most draw commands: the
// warmed-up presentation frame rather than an early allocation frame.
func busiestFrame(log string) string {
	sections := strings.Split(log, "Graphics commands:")

	busiest, most := "", 0
	for _, section := range sections {
		if draws := countDrawTriangles(section); draws > most {
			busiest, most = section, draws
		}
	}

	return busiest
}

// countDrawTriangles counts the draw commands in one ebitenginedebug frame.
func countDrawTriangles(log string) int {
	total := 0
	for line := range strings.SplitSeq(log, "\n") {
		if strings.Contains(line, "draw-triangles") {
			total++
		}
	}

	return total
}

// countCustomShaderDraws counts the draws through a compiled Kage program.
func countCustomShaderDraws(log string) int {
	return len(customShaderDraw.FindAllString(log, -1))
}

// clipWithLevels builds one Clip with its light levels pre-filled.
func clipWithLevels(x, y, width, height int, pixels, levels []byte) render.Clip {
	clone := make([]byte, len(pixels))
	copy(clone, pixels)

	lightLevels := make([]byte, len(levels))
	copy(lightLevels, levels)

	return render.Clip{X: x, Y: y, Width: width, Height: height, Pixels: clone, Light: lightLevels}
}

// applyFadeRGBA dims every expanded pixel by fadeval/256
// (ApplyFadeLevel, palette.cpp:140-147).
func applyFadeRGBA(pixels []render.RGBA, fadeval int) {
	for i := range pixels {
		pixels[i].R = uint8(fadeval * int(pixels[i].R) / fullFade)
		pixels[i].G = uint8(fadeval * int(pixels[i].G) / fullFade)
		pixels[i].B = uint8(fadeval * int(pixels[i].B) / fullFade)
	}
}

// expectVariantsTexture checks every texel of the texture upload against the
// variant row it claims to carry.
func expectVariantsTexture(t *testing.T, texture []byte, variants *render.Variants) {
	t.Helper()

	for shade, row := range variants {
		for index, rgb := range row {
			base := (shade*256 + index) * 4

			got := render.RGBA{R: texture[base], G: texture[base+1], B: texture[base+2], A: texture[base+3]}
			want := render.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 0xFF}

			if got != want {
				t.Fatalf("texture[%d][%d] = %v, want %v", shade, index, got, want)
			}
		}
	}
}

// expectFrameEqual compares two frames pixel by pixel.
func expectFrameEqual(t *testing.T, got, want render.Frame) {
	t.Helper()

	if got.Width != want.Width || got.Height != want.Height {
		t.Fatalf("frame size %dx%d, want %dx%d", got.Width, got.Height, want.Width, want.Height)
	}

	for i, g := range got.Pixels {
		if g != want.Pixels[i] {
			t.Fatalf("pixel %d = %v, want %v", i, g, want.Pixels[i])
		}
	}
}
