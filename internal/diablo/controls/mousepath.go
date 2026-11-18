package controls

// MousePathTarget is cursPosition (plrctrls.cpp:66 area): the dungeon tile the
// resolved cursor stands on — the target every mouse-path command names.
type MousePathTarget struct {
	// X, Y are the tile coordinates (cursPosition.x/.y).
	X, Y int
}

// MousePathInput is the pre-resolved target state LeftMouseCmd branches on
// (diablo.cpp:247-319 @ fork ca91d778f): the cursor ordinal, the level kind
// and the target ids under the cursor, all resolved by CheckCursMove
// (cursor.cpp:902-931) and its TrySelect* screens before the click. -1 id
// fields mean "no target". Explicit inputs, house PlayerView pattern — the
// adapter onto the live cursor state lives at the call site.
type MousePathInput struct {
	// Cursor is pcurs (cursor_defs.hpp; read at diablo.cpp:255,274).
	Cursor CursorID
	// Town is leveltype == DTYPE_TOWN (diablo.cpp:252,267).
	Town bool
	// Shift is bShift, the shift-held modifier (diablo.cpp:247).
	Shift bool
	// InventoryOpen is invflag — it flips the pickup command between the
	// near (GOTOGETITEM) and far (GOTOAGETITEM) variants (diablo.cpp:255-256).
	InventoryOpen bool
	// Item is pcursitem (diablo.cpp:255,271).
	Item int
	// Monster is pcursmonst (diablo.cpp:258,283,301).
	Monster int
	// Player is PlayerUnderCursor's id (diablo.cpp:290,313).
	Player int
	// Object is ObjectUnderCursor's id (diablo.cpp:271).
	Object int
	// Ranged is myPlayer.UsesRangedWeapon() (diablo.cpp:277).
	Ranged bool
}

// MousePickup is the SendPickup row's payload: the item to walk to and the
// near/far variant (diablo.cpp:255-256: CMD_GOTOGETITEM when invflag,
// CMD_GOTOAGETITEM otherwise).
type MousePickup struct {
	// Item is pcursitem, the item id.
	Item int
	// Near is true when the inventory is open (CMD_GOTOGETITEM).
	Near bool
}

// MouseObject is the SendOperateObject row's payload (diablo.cpp:271-275):
// the object to operate, disarmed when the disarm cursor is loaded.
type MouseObject struct {
	// Object is the object id under the cursor.
	Object int
	// Disarm is pcurs == CURSOR_DISARM (CMD_DISARMXY instead of CMD_OPOBJXY).
	Disarm bool
}

// MouseAttack is the SendAttack row's payload (diablo.cpp:277-317): the
// attack verb resolved to a tile, a monster id or a player id.
type MouseAttack struct {
	// Ranged is UsesRangedWeapon: CMD_RATTACK* over CMD_SATTACK*/ATTACKID.
	Ranged bool
	// Tile is the plain-position attack (SATTACKXY/RATTACKXY); when false,
	// Monster/Player name the id target.
	Tile bool
	// Monster is pcursmonst for the id attacks (CMD_ATTACKID/CMD_RATTACKID).
	Monster int
	// Player is the target player's id (CMD_ATTACKPID/CMD_RATTACKPID).
	Player int
}

// MousePathHooks are the environment functions the mouse path emits through —
// the NetSendCmd* port boundaries, exactly like StickHooks (network stays at
// the call site; the decision core never dials). A nil hook is skipped.
type MousePathHooks struct {
	// SendWalk ports NetSendCmdLoc(MyPlayerId, true, CMD_WALKXY, cursPosition)
	// (diablo.cpp:263,318).
	SendWalk func(x, y int)
	// SendTalk ports NetSendCmdLocParam1(true, CMD_TALKXY, cursPosition,
	// pcursmonst) (diablo.cpp:258-259).
	SendTalk func(monster int)
	// SendPickup ports NetSendCmdLocParam1(true, invflag ? CMD_GOTOGETITEM :
	// CMD_GOTOAGETITEM, cursPosition, pcursitem) (diablo.cpp:255-256).
	SendPickup func(MousePickup)
	// SendOperateObject ports NetSendCmdLoc(MyPlayerId, true,
	// pcurs == CURSOR_DISARM ? CMD_DISARMXY : CMD_OPOBJXY, cursPosition)
	// (diablo.cpp:272-274).
	SendOperateObject func(MouseObject)
	// SendAttack ports the attack family: CMD_RATTACKXY / CMD_SATTACKXY
	// (diablo.cpp:279-280,293-295,306-308) and the id attacks
	// (diablo.cpp:283-285,290-292,301-303,313-315).
	SendAttack func(MouseAttack)
}

// MousePathAction ports LeftMouseCmd (Source/diablo.cpp:247-319 @ fork
// ca91d778f): one mouse click over the world, resolved to the command rows
// its cursor and targets mean. Guard order is mirrored exactly; the town
// branch's three ifs are independent (diablo.cpp:255-260), so a click on a
// townsperson standing on an item both talks and picks up.
//
// Deviations, input-modeling only (no branch reordered, no branch cut):
//   - The upstream assert on main-panel containment (diablo.cpp:249) and the
//     CloseGoldWithdraw/CloseStash/CloseVisualStore calls (diablo.cpp:253-255)
//     are presentation/overlay state — the wiring layer owns them, the
//     decision rows below are untouched.
//   - The CanTalkToMonst nearness gate and the target liveness checks run
//     inside the real Send hooks' upstream callees; the decision rows here
//     are the verb choice LeftMouseCmd makes.
func MousePathAction(in MousePathInput, target MousePathTarget, hooks MousePathHooks) {
	if in.Town {
		mouseTown(in, target, hooks)
		return
	}
	mouseDungeon(in, target, hooks)
}

// mouseTown ports the town branch (diablo.cpp:252-265).
func mouseTown(in MousePathInput, target MousePathTarget, hooks MousePathHooks) {
	if in.Item != -1 && in.Cursor == CursorHand && hooks.SendPickup != nil {
		hooks.SendPickup(MousePickup{Item: in.Item, Near: in.InventoryOpen}) // diablo.cpp:255-256
	}
	if in.Monster != -1 && hooks.SendTalk != nil {
		hooks.SendTalk(in.Monster) // diablo.cpp:258-259
	}
	if in.Item == -1 && in.Monster == -1 && in.Player == -1 && hooks.SendWalk != nil {
		hooks.SendWalk(target.X, target.Y) // diablo.cpp:261-263
	}
}

// mouseDungeon ports the dungeon branch (diablo.cpp:267-318): the pickup and
// object rows, the ranged/melee split, and the walk fallthrough.
func mouseDungeon(in MousePathInput, target MousePathTarget, hooks MousePathHooks) {
	if in.Item != -1 && in.Cursor == CursorHand && !in.Shift && hooks.SendPickup != nil {
		hooks.SendPickup(MousePickup{Item: in.Item, Near: in.InventoryOpen}) // diablo.cpp:271-272
		return
	}
	if in.Object != -1 && hooks.SendOperateObject != nil {
		// diablo.cpp:273-276: enabled object; the disarm cursor picks the verb.
		hooks.SendOperateObject(MouseObject{Object: in.Object, Disarm: in.Cursor == CursorDisarm})
		return
	}
	if in.Ranged {
		mouseRanged(in, target, hooks)
	} else {
		mouseMelee(in, target, hooks)
	}
	mouseWalkFallthrough(in, target, hooks)
}

// mouseWalkFallthrough ports the dungeon walk fallthrough (diablo.cpp:317-319):
// an unshifted click with no target under the cursor walks to the tile.
func mouseWalkFallthrough(in MousePathInput, target MousePathTarget, hooks MousePathHooks) {
	if in.Shift || in.Item != -1 || in.Object != -1 || in.Monster != -1 || in.Player != -1 {
		return
	}
	if hooks.SendWalk != nil {
		hooks.SendWalk(target.X, target.Y) // diablo.cpp:317-319
	}
}

// mouseRanged ports the ranged rows (diablo.cpp:277-299).
func mouseRanged(in MousePathInput, _ MousePathTarget, hooks MousePathHooks) {
	if hooks.SendAttack == nil {
		return
	}
	if in.Shift {
		hooks.SendAttack(MouseAttack{Ranged: true, Tile: true}) // diablo.cpp:279-280
		return
	}
	if in.Monster != -1 {
		hooks.SendAttack(MouseAttack{Ranged: true, Monster: in.Monster}) // diablo.cpp:283-290
		return
	}
	if in.Player != -1 {
		hooks.SendAttack(MouseAttack{Ranged: true, Player: in.Player}) // diablo.cpp:290-292
	}
}

// mouseMelee ports the melee rows (diablo.cpp:300-316).
func mouseMelee(in MousePathInput, _ MousePathTarget, hooks MousePathHooks) {
	if hooks.SendAttack == nil {
		return
	}
	if in.Shift {
		hooks.SendAttack(MouseAttack{Tile: true}) // diablo.cpp:303-308
		return
	}
	if in.Monster != -1 {
		hooks.SendAttack(MouseAttack{Monster: in.Monster}) // diablo.cpp:310-311
		return
	}
	if in.Player != -1 {
		hooks.SendAttack(MouseAttack{Player: in.Player}) // diablo.cpp:313-315
	}
}
