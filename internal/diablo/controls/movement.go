package controls

// PlayerView is the narrow read surface the movement pipeline needs over
// entities.Player (house pattern: engine/path.go stepContext hooks — the
// pipeline takes explicit inputs so decision order is testable before the
// entity wiring lands). The adapter onto entities.Player lives at the
// call site; entities never imports controls.
type PlayerView interface {
	// Mode is _pmode (player.h:389): the current PLR_MODE.
	Mode() PlrMode
	// AnimFrame is AnimInfo.CurrentFrame (player.h:262).
	AnimFrame() int8
	// AttackAnimFrames is _pAFNum (player.h:264): the attack animation's
	// action frame.
	AttackAnimFrames() int8
	// SpellAnimFrames is _pSFNum (player.h:266): the spell animation's
	// action frame.
	SpellAnimFrames() int8
	// Facing is _pdir (player.h:391): the faced direction.
	Facing() Direction
	// Future is position.future (engine/actor_position.hpp:16): the tile the
	// actor's walk is moving into.
	Future() (int, int)
	// WalkpathHead is walkpath[0] (player.h:261): WalkNone (-1) when idle.
	WalkpathHead() int8
	// DestAction is destAction (player.h:263): DestActionNone (-1) when idle.
	DestAction() int8
	// SetFacing writes _pdir (plrctrls.cpp:1739,1752).
	SetFacing(Direction)
}

// StickInput is the raw per-frame stick state Movement reads (the upstream
// leftStickX/leftStickY/rightStickX/rightStickY globals plus the overlay
// guards, plrctrls.cpp:1834-1838): raw floats, no deadzone applied — the
// wiring lane scales axes before they get here.
type StickInput struct {
	// X/Y are the left stick axes.
	X, Y float64
	// RightX/RightY are the right stick axes (significance check only,
	// plrctrls.cpp:1891-1892).
	RightX, RightY float64
	// MenuNavigatorActive is PadMenuNavigatorActive (plrctrls.cpp:1834).
	MenuNavigatorActive bool
	// HotspellMenuActive is PadHotspellMenuActive (plrctrls.cpp:1834).
	HotspellMenuActive bool
	// InGameMenu is InGameMenu() (plrctrls.cpp:1834).
	InGameMenu bool
	// GameUIHandler mirrors GetLeftStickOrDPadGameUIHandler() != nullptr
	// (plrctrls.cpp:1837): a custom game-UI consumer owns the stick while
	// set. Nil = unset. The handler itself stays at the wiring layer.
	GameUIHandler func(AxisDirection)
}

// StickHooks are the environment functions the pipeline emits through: the
// NetSendCmdLoc port boundary (network command emission stays at the call
// site — the decision core must not dial the network) plus the tile-query
// and player-state hooks whose real implementations live with the level and
// player packages.
type StickHooks struct {
	// IsStandingGround ports IsStandingGround (plrctrls.cpp:438-444): the
	// StandGround padmapper toggle/combo check.
	IsStandingGround func() bool
	// StartStand ports StartStand (player's stand-state entry).
	StartStand func(Direction)
	// PosOkPlayer ports PosOkPlayer (motion.cpp): can the player occupy the
	// tile.
	PosOkPlayer func(x, y int) bool
	// IsTileNotSolid ports IsTileNotSolid (levels/tile lookup).
	IsTileNotSolid func(x, y int) bool
	// SendWalkCommand ports the CMD_WALKXY emission
	// (NetSendCmdLoc(plr.getId(), true, CMD_WALKXY, delta), plrctrls.cpp:1747).
	SendWalkCommand func(x, y int)
	// SendStopCommand ports the stop-walk CMD_WALKXY emission
	// (NetSendCmdLoc(..., player.position.future), plrctrls.cpp:1730).
	SendStopCommand func(x, y int)
}

// TurnToDir ports TurnToDir (plrctrls.cpp:1750-1754): a gentle stick
// deflection turns the player in place; a no-op while walking, while the
// action cannot change, or for NoDirection.
func TurnToDir(view PlayerView, dir Direction, hooks StickHooks) {
	if isWalking(view) || !canChangeAction(view) || dir == NoDirection {
		return
	}
	view.SetFacing(dir)
	hooks.StartStand(dir)
}

// IsPathBlocked ports IsPathBlocked (plrctrls.cpp:1704-1723): cardinal steps
// (North/East/South/West) never count as blocked; diagonal steps are blocked
// when both left/right corner steps fail PosOkPlayer (unless both corners
// are plain non-solid tiles, which passes without the PosOkPlayer check).
func IsPathBlocked(view PlayerView, dir Direction, hooks StickHooks) bool {
	if dir != North && dir != East && dir != South && dir != West {
		return false // Steps along a major axis don't need to check corners
	}

	futureX, futureY := view.Future()
	leftDX, leftDY := DirectionDelta(Left(dir))
	rightDX, rightDY := DirectionDelta(Right(dir))
	leftX, leftY := futureX+leftDX, futureY+leftDY
	rightX, rightY := futureX+rightDX, futureY+rightDY

	if hooks.IsTileNotSolid(leftX, leftY) && hooks.IsTileNotSolid(rightX, rightY) {
		return false
	}
	return !hooks.PosOkPlayer(leftX, leftY) && !hooks.PosOkPlayer(rightX, rightY)
}

// WalkInDir ports WalkInDir (plrctrls.cpp:1725-1747): a firm stick push
// stops the current walk (NONE axis) or walks one tile toward the pushed
// direction; standing-ground players only re-face.
func WalkInDir(view PlayerView, dir AxisDirection, hooks StickHooks) {
	if dir.X == AxisXNone && dir.Y == AxisYNone {
		stopWalking(view, hooks)
		return
	}

	pdir := FaceDir[dir.X][dir.Y]
	dx, dy := DirectionDelta(pdir)
	futureX, futureY := view.Future()
	deltaX, deltaY := futureX+dx, futureY+dy

	if !isWalking(view) && canChangeAction(view) {
		view.SetFacing(pdir)
	}

	if hooks.IsStandingGround() {
		standFacing(view, pdir, hooks)
		return
	}

	if hooks.PosOkPlayer(deltaX, deltaY) && IsPathBlocked(view, pdir, hooks) {
		standFacing(view, pdir, hooks)
		return // Don't start backtrack around obstacles
	}

	hooks.SendWalkCommand(deltaX, deltaY)
}

// Movement ports Movement (plrctrls.cpp:1832-1853): the per-frame left-stick
// entry point. Overlay consumers (menu navigator, hotspell menu, in-game
// menu, custom game-UI handler) take the stick before movement runs; a firm
// push walks, a gentle push turns in place.
func Movement(view PlayerView, input StickInput, hooks StickHooks) {
	if input.MenuNavigatorActive || input.HotspellMenuActive || input.InGameMenu {
		return
	}
	if input.GameUIHandler != nil {
		return
	}

	// Gentle deflection turns in place, a firm push walks.
	dir := StickDirection(input.X, input.Y)
	pdir := FaceDir[dir.X][dir.Y]
	switch ClassifyStickPush(input.X, input.Y) {
	case PushFirm:
		WalkInDir(view, dir, hooks)
	case PushWeak:
		TurnToDir(view, pdir, hooks)
	default: // StickPush::None
	}
}

// IsStickMovementSignificant ports IsStickMovementSignificant
// (plrctrls.cpp:1890-1893): a firm left-stick push or any right-stick
// deflection counts as stick movement (keeps the auto-hints visible).
func IsStickMovementSignificant(input StickInput) bool {
	return ClassifyStickPush(input.X, input.Y) == PushFirm ||
		input.RightX != 0 || input.RightY != 0
}

// stopWalking is WalkInDir's NONE-axis branch (plrctrls.cpp:1726-1731): stop
// an active walkpath with no pending dest action by issuing CMD_WALKXY to
// position.future. KeyboardAndMouse mode keeps walkpath control upstream;
// the ported pipeline is gamepad-only so the ControlMode check is
// structural-true here (the wiring lane never calls WalkInDir for KB&M).
func stopWalking(view PlayerView, hooks StickHooks) {
	if view.WalkpathHead() != WalkNone && view.DestAction() == DestActionNone {
		x, y := view.Future()
		hooks.SendStopCommand(x, y) // Stop walking
	}
}

// standFacing is WalkInDir's shared StartStand gate
// (plrctrls.cpp:1741-1744,1749-1752): only a standing player re-enters its
// stand state; a walking player neither stands nor proceeds.
func standFacing(view PlayerView, pdir Direction, hooks StickHooks) {
	if view.Mode() == ModeStand {
		hooks.StartStand(pdir)
	}
}

// isWalking ports Player::isWalking (player.cpp:1709-1713): the PM_WALK_*
// trio; the Go entity enum collapses them to ModeWalk.
func isWalking(view PlayerView) bool {
	return view.Mode() == ModeWalk
}

// canChangeAction ports Player::CanChangeAction (player.h:761-774): stand
// always passes; attack/rattack pass from _pAFNum on; spell from _pSFNum on.
// The upstream isWalking && isLastFrame branch (player.h:771-772) is not
// ported: both ported call sites early-return on isWalking() before
// consulting this guard, so the branch is unreachable here (documented
// deviation, see lane log).
func canChangeAction(view PlayerView) bool {
	switch view.Mode() {
	case ModeStand:
		return true
	case ModeAttack:
		return view.AnimFrame() >= view.AttackAnimFrames()
	case ModeSpell:
		return view.AnimFrame() >= view.SpellAnimFrames()
	default:
		return false
	}
}
