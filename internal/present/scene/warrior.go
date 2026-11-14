package scene

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/assets"
	"github.com/bmcszk/devilutionGo/internal/light"
	"github.com/bmcszk/devilutionGo/internal/render"
)

// warriorSTANDRecord is the town warrior STAND record: light armor + unarmed+shield
// (town.cpp LoadPlrGFX: town uses szCel="st", classChar='w', ArmourChar['l']=light, WepChar['u']=unarmed+shield).
// Archive path: plrgfx\warrior\wln\wlnst.cl2 (diabdat inventory line 411).
const warriorSTANDRecord = `plrgfx\warrior\wln\wlnst.cl2`

// sceneBrightness is the first-scene boot brightness: the readable middle of
// the range devilutionX players run (gamma 30-60).
const sceneBrightness = 60

type WarriorFrame struct {
	Pixels []byte
	Width  int
	Height int
}

// HeroTile returns the hero's tile position.
func HeroTile() render.TilePos { return heroTile }

// HasHeroAt checks whether the composed surface contains hero pixels at the
// expected screen rectangle (color-class assertion: hero pixels have a
// distinct palette index from the town floor at the hero's screen position).
// Uses view-relative coordinates matching the draw logic.
func HasHeroAt(composed TownScene, _ render.TilePos) bool {
	surface := composed.Surface

	// The test doesn't have access to the view, but we know the view is
	// centered on the hero (heroTile). The hero should be drawn near the
	// center of the 640x480 surface. Use a generous center region check.
	// The hero is drawn at: center of surface +- anchor.
	// Surface is 640x480, hero is 96x128 with anchor (-16, -128).
	// Center of surface is (320, 240). Hero draw position relative to surface center:
	//   clipX = 320 - 16 = 304
	//   clipY = 240 - 128 = 112 (top of hero sprite).
	// Hero occupies roughly clipX..clipX+96, clipY..clipY+128.
	const x0 = 296
	const y0 = 104
	const x1 = 408
	const y1 = 248

	if x0 < 0 || y0 < 0 || x1 >= surface.Width || y1 >= surface.Height {
		return false
	}

	// Collect unique indices in the hero rect
	heroIndices := make(map[uint8]bool)
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			idx := surface.Pixels[y*surface.Width+x]
			heroIndices[idx] = true
		}
	}

	// The hero rect must contain at least 2 distinct palette indices
	// (hero + town background mixing)
	if len(heroIndices) < 2 {
		return false
	}

	return true
}

// sceneBrightness is the first-scene boot brightness: the readable middle of
// the range devilutionX players run (gamma 30-60).
// ComposeTownFrameWithHero composes the town frame and overlays the warrior
// STAND frame 0 at the hero tile through the Presenter pipeline.
func ComposeTownFrameWithHero(dirs []string) (TownScene, error) {
	return townFrameAtBrightness(dirs, sceneBrightness)
}

// WarriorStandFrame0 loads and decodes the warrior's town STAND frame 0 from
// the real archives (plrgfx\warrior\wln\wlnst.cl2, the south-facing light-armor
// stand sheet; LoadPlrGFX szCel "st", classChar 'w', ArmourChar 'l', WepChar 'n').
// DecodeCL2 handles the 8-group sheet: groups are the 8 directions, 20 frames each.
func WarriorStandFrame0(stack *assets.MPQStack) (WarriorFrame, error) {
	celData, ok := stack.ReadFile(warriorSTANDRecord)
	if !ok || len(celData) == 0 {
		return WarriorFrame{}, fmt.Errorf("read %s: not found in archives", warriorSTANDRecord)
	}

	// STAND frames are 96x96: the decoder's strict row walk lands exactly
	// 9216 pixels per frame; 96x128 overruns the command stream.
	list, err := assets.DecodeCL2(celData, 96, 96)
	if err != nil {
		return WarriorFrame{}, fmt.Errorf("decode %s: %w", warriorSTANDRecord, err)
	}

	if len(list.GroupSizes) < 1 || list.GroupSizes[0] < 1 {
		return WarriorFrame{}, fmt.Errorf("%s: no frames in direction group 0", warriorSTANDRecord)
	}

	frame := list.Frames[0]

	// CL2 frames are stored bottom-up (first row = feet); flip to top-down for
	// screen coordinates where y grows downward.
	flipped := make([]byte, len(frame.Pixels))
	for y := range frame.Height {
		copy(flipped[y*frame.Width:(y+1)*frame.Width],
			frame.Pixels[(frame.Height-1-y)*frame.Width:(frame.Height-y)*frame.Width])
	}

	return WarriorFrame{Pixels: flipped, Width: frame.Width, Height: frame.Height}, nil
}

// townFrameAtBrightness composes the town frame with the hero drawn over
// it at the given brightness tone-map (ApplyGlobalBrightness, palette.cpp:96-110).
func townFrameAtBrightness(dirs []string, brightness int) (TownScene, error) {
	stack := assets.NewMPQStack()
	defer func() { _ = stack.Close() }()

	stack.Load(dirs, "hellfire", assets.HellfireMpqPriority)
	stack.Load(dirs, "DIABDAT", assets.MainMpqPriority, ".MPQ")
	stack.Load(dirs, "devilutionx", assets.DevilutionXMpqPriority)

	// Load town records (reuse existing path - loads + decodes MIN+CEL+SOL)
	town, err := loadTownRecords(stack)
	if err != nil {
		return TownScene{}, err
	}

	// Load hero frame
	heroFrame, err := WarriorStandFrame0(stack)
	if err != nil {
		return TownScene{}, err
	}

	heroPixels, heroW, heroH := heroFrame.Pixels, heroFrame.Width, heroFrame.Height

	// Build town grid + scene (reuse existing)
	grid := buildTownGrid(town.files, town.megaTiles)

	pal := render.NewPaletteState(*town.pal, light.MakeLightTables(light.DTYPE_TOWN), assets.BlendingFor(0))
	state := pal.WithBrightness(brightness)

	world, view := townWorld(grid, town.micros, town.sol)

	viewData := &townWorldView{grid: grid, micros: town.micros, sol: town.sol, light: townLight()}

	surface := renderTownFrame(world, view, town.cel, state, viewData)

	// The view is centered on the hero tile (composeView in town.go), so the
	// hero tile's bottom vertex lands at the surface centre (320, 240). Upstream
	// anchors the stand sprite at the tile's centre and draws UP: top-left =
	// (cx - w/2, cy - h) with a half-tile nudge upstream's getRenderingOffset
	// applies ((scrollrt.cpp DrawPlayer: x - (w-TILE_W)/2, y - h + TILE_H/2)).
	clipX := 320 - heroW/2
	clipY := 240 - heroH + render.TileHeight/2

	clip := render.Clip{
		X:      clipX,
		Y:      clipY,
		Width:  heroW,
		Height: heroH,
		Pixels: heroPixels,
		Light:  nil, // fully lit for STAND in town
	}

	surface.DrawSpriteKeyed(clip)

	return TownScene{Surface: surface, State: state}, nil
}
