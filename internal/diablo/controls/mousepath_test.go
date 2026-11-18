// Package controls ports the mouse click-to-walk path: LeftMouseCmd's town
// and dungeon verb decision (Source/diablo.cpp:247-319 @ fork ca91d778f).
package controls_test

import (
	"testing"

	"github.com/bmcszk/devilutionGo/internal/diablo/controls"
)

// mouseVerbCases pin LeftMouseCmd's verb choice per input row
// (diablo.cpp:247-319): town rows, dungeon rows, the walk fallthroughs.
var mouseVerbCases = []struct {
	name  string
	input controls.MousePathInput
	want  string
}{
	{
		name:  "town plain click walks",
		input: defaultInput(controls.MousePathInput{Town: true}),
		want:  "walk",
	},
	{
		name:  "town click on a towner talks",
		input: defaultInput(controls.MousePathInput{Town: true, Monster: 7}),
		want:  "talk",
	},
	{
		name:  "town click on an item picks it up",
		input: defaultInput(controls.MousePathInput{Town: true, Item: 3}),
		want:  "pickup",
	},
	{
		name:  "town item with hand cursor and open inventory asks the near pickup",
		input: defaultInput(controls.MousePathInput{Town: true, Item: 3, InventoryOpen: true}),
		want:  "pickup",
	},
	{
		// diablo.cpp:261-263: the walk needs ALL of item/monster/player
		// clear, so an item the cursor cannot lift leaves the click inert.
		name:  "town item under an item cursor fires nothing",
		input: defaultInput(controls.MousePathInput{Town: true, Item: 3, Cursor: controls.CursorTeleport}),
		want:  "nothing",
	},
	{
		name:  "town shift click still walks",
		input: defaultInput(controls.MousePathInput{Town: true, Shift: true}),
		want:  "walk",
	},
	{
		name:  "dungeon plain click walks",
		input: defaultInput(controls.MousePathInput{}),
		want:  "walk",
	},
	{
		name:  "dungeon shift click attacks the tile",
		input: defaultInput(controls.MousePathInput{Shift: true}),
		want:  "attack",
	},
	{
		name:  "dungeon item click picks up",
		input: defaultInput(controls.MousePathInput{Item: 5}),
		want:  "pickup",
	},
	{
		name:  "dungeon ranged weapon shift attacks the tile",
		input: defaultInput(controls.MousePathInput{Shift: true, Ranged: true}),
		want:  "attack",
	},
}

// mouseWorld is the recorded hook calls: the verb rows MousePathAction fired.
type mouseWorld struct {
	talks   []int
	pickups []controls.MousePickup
	objects []controls.MouseObject
	attacks []controls.MouseAttack
	walks   [][2]int
}

func (w *mouseWorld) hooks() controls.MousePathHooks {
	return controls.MousePathHooks{
		SendTalk:          func(monster int) { w.talks = append(w.talks, monster) },
		SendPickup:        func(p controls.MousePickup) { w.pickups = append(w.pickups, p) },
		SendOperateObject: func(o controls.MouseObject) { w.objects = append(w.objects, o) },
		SendAttack:        func(a controls.MouseAttack) { w.attacks = append(w.attacks, a) },
		SendWalk:          func(x, y int) { w.walks = append(w.walks, [2]int{x, y}) },
	}
}

// fired reports the single verb the world recorded, or an error description.
func (w *mouseWorld) fired() string {
	total := len(w.talks) + len(w.pickups) + len(w.objects) + len(w.attacks) + len(w.walks)
	switch {
	case total != 1:
		if total == 0 {
			return "nothing"
		}
		return "not exactly one hook fired"
	case len(w.talks) == 1:
		return "talk"
	case len(w.pickups) == 1:
		return "pickup"
	case len(w.objects) == 1:
		return "object"
	case len(w.attacks) == 1:
		return "attack"
	default:
		return "walk"
	}
}

// TestMousePathAction dispatches each verb case row through MousePathAction.
func TestMousePathAction(t *testing.T) {
	for _, tc := range mouseVerbCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			tile := controls.MousePathTarget{X: 66, Y: 68}
			w := &mouseWorld{}

			// when
			controls.MousePathAction(tc.input, tile, w.hooks())

			// then
			if got := w.fired(); got != tc.want {
				t.Fatalf("MousePathAction(%+v) fired %s, want %s (world %+v)", tc.input, got, tc.want, w)
			}
		})
	}
}

// TestMousePathTownWalkTarget pins the town walk payload: CMD_WALKXY to
// cursPosition (diablo.cpp:263).
func TestMousePathTownWalkTarget(t *testing.T) {
	// given
	w := &mouseWorld{}

	// when
	controls.MousePathAction(
		defaultInput(controls.MousePathInput{Town: true}),
		controls.MousePathTarget{X: 66, Y: 68},
		w.hooks(),
	)

	// then
	assertWalk(t, w, 66, 68)
}

// TestMousePathPickupVerb pins the pickup verb's near/far split: the open
// inventory asks CMD_GOTOGETITEM (the near pickup) instead of
// CMD_GOTOAGETITEM (diablo.cpp:255-256).
func TestMousePathPickupVerb(t *testing.T) {
	cases := []struct {
		name      string
		inventory bool
		wantNear  bool
	}{
		{name: "closed inventory asks the far pickup", inventory: false, wantNear: false},
		{name: "open inventory asks the near pickup", inventory: true, wantNear: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			w := &mouseWorld{}

			// when
			controls.MousePathAction(defaultInput(controls.MousePathInput{
				Town:          true,
				Item:          3,
				Cursor:        controls.CursorHand,
				InventoryOpen: tc.inventory,
			}), controls.MousePathTarget{X: 1, Y: 2}, w.hooks())

			// then
			if len(w.pickups) != 1 || w.pickups[0].Item != 3 || w.pickups[0].Near != tc.wantNear {
				t.Fatalf("pickups = %+v, want one pickup of item 3 near=%v", w.pickups, tc.wantNear)
			}
		})
	}
}

// TestMousePathObject pins the dungeon object branch (diablo.cpp:273-276):
// the hand cursor operates with CMD_OPOBJXY, the disarm cursor sends
// CMD_DISARMXY instead.
func TestMousePathObject(t *testing.T) {
	cases := []struct {
		name       string
		cursor     controls.CursorID
		wantDisarm bool
	}{
		{name: "hand cursor operates", cursor: controls.CursorHand, wantDisarm: false},
		{name: "disarm cursor disarms", cursor: controls.CursorDisarm, wantDisarm: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			w := &mouseWorld{}

			// when
			controls.MousePathAction(
				defaultInput(controls.MousePathInput{Object: 9, Cursor: tc.cursor}),
				controls.MousePathTarget{X: 3, Y: 4},
				w.hooks(),
			)

			// then
			if len(w.objects) != 1 || w.objects[0].Object != 9 || w.objects[0].Disarm != tc.wantDisarm {
				t.Fatalf("objects = %+v, want one object row for 9 disarm=%v", w.objects, tc.wantDisarm)
			}
		})
	}
}

// TestMousePathAttack pins the dungeon attack branches: melee shift fires
// the tile attack (SATTACKXY), ranged shift the ranged tile attack
// (RATTACKXY), and an unshifted target the id attacks (diablo.cpp:276-313).
func TestMousePathAttack(t *testing.T) {
	tile := controls.MousePathTarget{X: 5, Y: 6}

	cases := []struct {
		name       string
		input      controls.MousePathInput
		wantRanged bool
		wantTile   bool
	}{
		{
			name:     "shift melee fires the tile attack",
			input:    defaultInput(controls.MousePathInput{Shift: true}),
			wantTile: true,
		},
		{
			name:       "shift ranged fires the ranged tile attack",
			input:      defaultInput(controls.MousePathInput{Shift: true, Ranged: true}),
			wantRanged: true,
			wantTile:   true,
		},
		{
			name:  "monster target id attack",
			input: defaultInput(controls.MousePathInput{Monster: 11}),
		},
		{
			name:       "ranged monster target fires the ranged id attack",
			input:      defaultInput(controls.MousePathInput{Monster: 11, Ranged: true}),
			wantRanged: true,
		},
		{
			name:  "player target id attack",
			input: defaultInput(controls.MousePathInput{Player: 2}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			w := &mouseWorld{}

			// when
			controls.MousePathAction(tc.input, tile, w.hooks())

			// then
			if got := w.fired(); got != "attack" {
				t.Fatalf("fired %s, want attack", got)
			}
			assertAttackWords(
				t,
				w.attacks[0],
				controls.MouseAttack{Ranged: tc.wantRanged, Tile: tc.wantTile},
				controls.MouseAttack{Tile: tc.wantTile},
			)
			assertAttackTarget(t, w.attacks[0], tc.input)
		})
	}
}

// TestMousePathTownMonsterAndItem pins the town guard order: the item and
// towner ifs are independent upstream (diablo.cpp:255-260), so a click on a
// townsperson standing on an item both talks and picks up.
func TestMousePathTownMonsterAndItem(t *testing.T) {
	// given
	w := &mouseWorld{}

	// when
	controls.MousePathAction(
		defaultInput(controls.MousePathInput{Town: true, Item: 3, Monster: 7}),
		controls.MousePathTarget{X: 1, Y: 2},
		w.hooks(),
	)

	// then
	if len(w.talks) != 1 || w.talks[0] != 7 {
		t.Fatalf("talks = %v, want one talk to monster 7", w.talks)
	}
	if len(w.pickups) != 1 || w.pickups[0].Item != 3 {
		t.Fatalf("pickups = %+v, want the item pickup too (independent ifs upstream)", w.pickups)
	}
	if len(w.walks) != 0 {
		t.Fatalf("walks = %v, want none when both targets matched", w.walks)
	}
}

// TestMousePathObjectGuardChain pins the dungeon object row's guard chain
// (diablo.cpp:271-273): disabled objects refuse, shift restricts to
// breakable objects in near range, and an item match suppresses the object
// row entirely (the pickup fires alone).
func TestMousePathObjectGuardChain(t *testing.T) {
	cases := []struct {
		name        string
		in          controls.MousePathInput
		wantObjects int
		wantPickups int
	}{
		{
			name:        "disabled object refuses",
			in:          defaultInput(controls.MousePathInput{Object: 4, ObjectDisabled: true}),
			wantObjects: 0,
		},
		{
			name:        "enabled object operates",
			in:          defaultInput(controls.MousePathInput{Object: 4}),
			wantObjects: 1,
		},
		{
			name:        "shift near breakable object operates",
			in:          defaultInput(controls.MousePathInput{Object: 4, Shift: true, Near: true, ObjectBreakable: true}),
			wantObjects: 1,
		},
		{
			name:        "shift near non-breakable object refuses",
			in:          defaultInput(controls.MousePathInput{Object: 4, Shift: true, Near: true}),
			wantObjects: 0,
		},
		{
			name:        "shift far breakable object refuses",
			in:          defaultInput(controls.MousePathInput{Object: 4, Shift: true, ObjectBreakable: true}),
			wantObjects: 0,
		},
		{
			name:        "item and object under cursor pickup only",
			in:          defaultInput(controls.MousePathInput{Item: 2, Object: 4}),
			wantObjects: 0,
			wantPickups: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			w := &mouseWorld{}

			// when
			controls.MousePathAction(tc.in, controls.MousePathTarget{X: 3, Y: 4}, w.hooks())

			// then
			if len(w.objects) != tc.wantObjects || len(w.pickups) != tc.wantPickups {
				t.Fatalf("objects = %+v (%d), pickups = %+v (%d), want objects=%d pickups=%d",
					w.objects, len(w.objects), w.pickups, len(w.pickups), tc.wantObjects, tc.wantPickups)
			}
		})
	}
}

// TestMousePathTalkVerbUnderRanged pins CanTalkToMonst's verb split
// (diablo.cpp:279-285, :293-294): a talkable monster still routes to the
// melee ATTACKID verb — Ranged stays the weapon truth and Talk names the
// verb; the melee gate lives on the shift row only.
func TestMousePathTalkVerbUnderRanged(t *testing.T) {
	cases := []struct {
		name string
		in   controls.MousePathInput
		want controls.MouseAttack
	}{
		{
			name: "ranged talkable monster fires melee ATTACKID",
			in:   defaultInput(controls.MousePathInput{Monster: 6, Ranged: true, CanTalkToMonst: true}),
			want: controls.MouseAttack{Ranged: true, Talk: true},
		},
		{
			name: "ranged non-talkable monster fires RATTACKID",
			in:   defaultInput(controls.MousePathInput{Monster: 6, Ranged: true}),
			want: controls.MouseAttack{Ranged: true},
		},
		{
			// diablo.cpp:291-294: the melee shift row carries the
			// CanTalkToMonst gate — talkable -> CMD_ATTACKID with the verb word.
			name: "melee shift talkable monster fires ATTACKID",
			in:   defaultInput(controls.MousePathInput{Monster: 6, Shift: true, CanTalkToMonst: true}),
			want: controls.MouseAttack{Talk: true},
		},
		{
			// diablo.cpp:295-302: melee shift without a talkable monster is
			// the plain tile attack (CMD_SATTACKXY).
			name: "melee shift non-talkable monster fires the tile attack",
			in:   defaultInput(controls.MousePathInput{Monster: 6, Shift: true}),
			want: controls.MouseAttack{Tile: true},
		},
		{
			// diablo.cpp:303-305: the melee unshifted monster row sends
			// plain CMD_ATTACKID with no CanTalkToMonst gate and no talk word.
			name: "melee unshifted monster fires plain ATTACKID",
			in:   defaultInput(controls.MousePathInput{Monster: 6}),
			want: controls.MouseAttack{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			w := &mouseWorld{}

			// when
			controls.MousePathAction(tc.in, controls.MousePathTarget{X: 5, Y: 6}, w.hooks())

			// then
			assertTalkRow(t, w.attacks, tc.want)
		})
	}
}

// TestMousePathPlayerTargetRefusals pins the hasNoLife/friendlyMode
// exclusions on both verb families (diablo.cpp:286, :306): a dead or
// friendly-mode player target refuses — no attack fires and no walk
// fallthrough takes the click.
func TestMousePathPlayerTargetRefusals(t *testing.T) {
	cases := []struct {
		name string
		in   controls.MousePathInput
	}{
		{
			name: "ranged dead player refuses",
			in:   defaultInput(controls.MousePathInput{Player: 2, Ranged: true, PlayerDead: true}),
		},
		{
			name: "ranged friendly mode refuses",
			in:   defaultInput(controls.MousePathInput{Player: 2, Ranged: true, FriendlyMode: true}),
		},
		{
			name: "melee dead player refuses",
			in:   defaultInput(controls.MousePathInput{Player: 2, PlayerDead: true}),
		},
		{
			name: "melee friendly mode refuses",
			in:   defaultInput(controls.MousePathInput{Player: 2, FriendlyMode: true}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			w := &mouseWorld{}

			// when
			controls.MousePathAction(tc.in, controls.MousePathTarget{X: 5, Y: 6}, w.hooks())

			// then
			if got := w.fired(); got != "nothing" {
				t.Fatalf("fired %s, want nothing (guard refusal) (world %+v)", got, w)
			}
		})
	}
}

// defaultInput is the no-target click: every id sentinel at -1, the hand
// cursor loaded (upstream's pcursitem/pcursmonst init at -1,
// plrctrls.cpp:2287-2290, and CURSOR_HAND as pcurs's resting value).
func defaultInput(in controls.MousePathInput) controls.MousePathInput {
	if in.Cursor == 0 {
		in.Cursor = controls.CursorHand
	}
	if in.Item == 0 {
		in.Item = -1
	}
	if in.Monster == 0 {
		in.Monster = -1
	}
	if in.Player == 0 {
		in.Player = -1
	}
	if in.Object == 0 {
		in.Object = -1
	}
	return in
}

// assertWalk asserts exactly one walk to the given tile.
func assertWalk(t *testing.T, w *mouseWorld, wantX, wantY int) {
	t.Helper()
	if len(w.walks) != 1 || w.walks[0][0] != wantX || w.walks[0][1] != wantY {
		t.Fatalf("walks = %v, want one walk to (%d,%d)", w.walks, wantX, wantY)
	}
}

// assertAttackWords asserts the attack row's ranged and tile words.
func assertAttackWords(t *testing.T, a, ranged, tile controls.MouseAttack) {
	t.Helper()
	if a.Ranged != ranged.Ranged || a.Tile != tile.Tile {
		t.Fatalf("attack = %+v, want ranged=%v tile=%v", a, ranged.Ranged, tile.Tile)
	}
}

// assertAttackTarget asserts an id attack names the input's monster or player.
func assertAttackTarget(t *testing.T, a controls.MouseAttack, in controls.MousePathInput) {
	t.Helper()
	if a.Tile {
		return
	}
	if in.Monster != -1 && a.Monster != in.Monster {
		t.Fatalf("attack = %+v, want monster=%d", a, in.Monster)
	}
	if in.Player != -1 && a.Player != in.Player {
		t.Fatalf("attack = %+v, want player=%d", a, in.Player)
	}
}

// assertTalkRow asserts exactly one attack hook fired with the verb words
// and the monster id target 6. Tile rows carry no id upstream
// (CMD_SATTACKXY sends the position only), so the id assert is skipped for
// them.
func assertTalkRow(t *testing.T, attacks []controls.MouseAttack, want controls.MouseAttack) {
	t.Helper()
	if len(attacks) != 1 {
		t.Fatalf("attacks = %+v, want exactly one", attacks)
	}
	a := attacks[0]
	if a.Ranged != want.Ranged || a.Talk != want.Talk || a.Tile != want.Tile {
		t.Fatalf("attack = %+v, want ranged=%v talk=%v tile=%v", a, want.Ranged, want.Talk, want.Tile)
	}
	if !want.Tile && a.Monster != 6 {
		t.Fatalf("attack = %+v, want monster=6", a)
	}
}
