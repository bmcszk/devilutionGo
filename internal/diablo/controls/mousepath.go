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
	// Near is myPlayer.position.tile.WalkingDistance(cursPosition) < 2
	// (diablo.cpp:269) — the shift+object row's near-range gate.
	Near bool
	// ObjectDisabled is ObjectUnderCursor->IsDisabled() (diablo.cpp:271).
	ObjectDisabled bool
	// ObjectBreakable is ObjectUnderCursor->_oBreak == 1 (diablo.cpp:271).
	ObjectBreakable bool
	// CanTalkToMonst is CanTalkToMonst(Monsters[pcursmonst]) (diablo.cpp:280,
	// :293) — it routes an id attack to the melee ATTACKID verb.
	CanTalkToMonst bool
	// PlayerDead is PlayerUnderCursor->hasNoLife() (diablo.cpp:290, :313).
	PlayerDead bool
	// FriendlyMode is myPlayer.friendlyMode (diablo.cpp:290, :313).
	FriendlyMode bool
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
	// Talk is the CanTalkToMonst branch: the melee ATTACKID verb, reachable
	// under a ranged weapon too (diablo.cpp:280-281, :293-294). When set,
	// Ranged stays the weapon truth and Talk names the verb.
	Talk bool
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
	// (diablo.cpp:277-278,296-297,299-301,306-308) and the id attacks
	// (diablo.cpp:280-285,293-294,303-305,306-308).
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
//   - The CanTalkToMonst nearness gate and the target liveness checks are
//     explicit inputs (Near, CanTalkToMonst, PlayerDead, FriendlyMode,
//     ObjectDisabled, ObjectBreakable) resolved by the caller — the
//     decision rows below are the verb choice LeftMouseCmd makes.
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

// mouseDungeon ports the dungeon branch (diablo.cpp:267-318): the pickup
// row, the object row of the if/else-if chain, the ranged/melee split, and
// the walk fallthrough. The item match suppresses the object row: upstream
// is one if/else-if chain (diablo.cpp:271-273).
func mouseDungeon(in MousePathInput, target MousePathTarget, hooks MousePathHooks) {
	if in.Item != -1 && in.Cursor == CursorHand && !in.Shift && hooks.SendPickup != nil {
		hooks.SendPickup(MousePickup{Item: in.Item, Near: in.InventoryOpen}) // diablo.cpp:271-272
		return
	}
	// diablo.cpp:273: enabled object; shift restricts to breakable objects in
	// near range (diablo.cpp:269 bNear). A refused object falls through to
	// the attack rows like upstream's else-if, never an invented no-op row.
	if objectOperable(in) && hooks.SendOperateObject != nil {
		hooks.SendOperateObject(MouseObject{Object: in.Object, Disarm: in.Cursor == CursorDisarm}) // diablo.cpp:273-276
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

// objectOperable is the object row's guard chain (diablo.cpp:271-273):
// the cursor is over an enabled object, and a shift click additionally
// needs near range and a breakable object (bNear from diablo.cpp:269).
func objectOperable(in MousePathInput) bool {
	if in.Object == -1 || in.ObjectDisabled {
		return false
	}
	return !in.Shift || (in.Near && in.ObjectBreakable)
}

// mouseRanged ports the ranged rows (diablo.cpp:276-288). The verb is
// CanTalkToMonst's ATTACKID even under a ranged weapon (diablo.cpp:279-281);
// dead and friendly-mode players refuse (diablo.cpp:286).
func mouseRanged(in MousePathInput, _ MousePathTarget, hooks MousePathHooks) {
	if hooks.SendAttack == nil {
		return
	}
	if in.Shift {
		hooks.SendAttack(MouseAttack{Ranged: true, Tile: true}) // diablo.cpp:279-280
		return
	}
	if in.Monster != -1 {
		if in.CanTalkToMonst { // diablo.cpp:280-281
			hooks.SendAttack(MouseAttack{Ranged: true, Talk: true, Monster: in.Monster})
			return
		}
		hooks.SendAttack(MouseAttack{Ranged: true, Monster: in.Monster}) // diablo.cpp:282-285
		return
	}
	if in.Player != -1 && !in.PlayerDead && !in.FriendlyMode { // diablo.cpp:286
		hooks.SendAttack(MouseAttack{Ranged: true, Player: in.Player}) // diablo.cpp:286-288
	}
}

// mouseMelee ports the melee rows (diablo.cpp:290-309): the shift split of
// CanTalkToMonst's ATTACKID (diablo.cpp:293-294) against the tile attack
// (diablo.cpp:295-302), the gateless unshifted monster row
// (diablo.cpp:303-305), and the hasNoLife/friendlyMode player exclusion
// (diablo.cpp:306).
func mouseMelee(in MousePathInput, _ MousePathTarget, hooks MousePathHooks) {
	if hooks.SendAttack == nil {
		return
	}
	if in.Shift && in.Monster != -1 && in.CanTalkToMonst { // diablo.cpp:291-294
		hooks.SendAttack(MouseAttack{Talk: true, Monster: in.Monster})
		return
	}
	if in.Shift {
		hooks.SendAttack(MouseAttack{Tile: true}) // diablo.cpp:295-302
		return
	}
	if in.Monster != -1 {
		hooks.SendAttack(MouseAttack{Monster: in.Monster}) // diablo.cpp:303-305
		return
	}
	if in.Player != -1 && !in.PlayerDead && !in.FriendlyMode { // diablo.cpp:306
		hooks.SendAttack(MouseAttack{Player: in.Player}) // diablo.cpp:306-308
	}
}
