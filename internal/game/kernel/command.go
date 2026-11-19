// Package kernel holds the deterministic simulation kernel: typed commands in,
// one ordered pass per tick, typed events out (ADR-015, ADR-018).
//
// The laws this package owns:
//
//	Law 1: everything that changes the game sends a Command carrying its tick.
//	Law 2: every real state change emits an Event; no change, no event.
//	Law 3: one ordered pass per tick, handlers run in registration order.
//	Law 6: the Gate drops commands that cannot change anything, before dispatch.
//
// The kernel never reads a clock, never starts a goroutine, and never lets map
// iteration decide an order: handler chains are slices in registration order
// and RNG streams are kept in sorted name order.
package kernel

// Kinds. Commands are domain + verb; the events they raise are past tense.
const (
	KindPlayerMove    Kind = "PlayerMove"
	KindPlayerAttack  Kind = "PlayerAttack"
	KindMouseMotion   Kind = "MouseMotion"
	KindMouseButton   Kind = "MouseButton"
	KindMouseWheel    Kind = "MouseWheel"
	KindKey           Kind = "Key"
	KindQuit          Kind = "Quit"
	KindMonsterDied   Kind = "MonsterDied"
	KindPlayerMoved   Kind = "PlayerMoved"
	KindPlayerStruck  Kind = "PlayerStruck"
	KindPointerMoved  Kind = "PointerMoved"
	KindGameQuit      Kind = "GameQuit"
	KindTickProcessed Kind = "TickProcessed"
	// KindDebugJumpLevel is the debug warp cheat (warp.cpp:18-30); its A word
	// carries the dungeon level.
	KindDebugJumpLevel Kind = "DebugJumpLevel"

	// Input bridge, stick family — the fork stick decision cores' outputs
	// (controller_motion.{h,cpp}, plrctrls.cpp Movement).
	// KindStickPush carries one classified left-stick decision: A is the
	// StickPush severity (controls.Push* ordinals), B and C the resolved
	// AxisDirectionX/Y sectors (controls.Axis* ordinals).
	KindStickPush Kind = "StickPush"

	// Input bridge, action family — the PerformPrimaryAction dispatch rows
	// (plrctrls.cpp:2309-2353). The kinds name the row the decision took; the
	// gate owns the refusals, the handlers own the effects.
	// KindPrimaryActionPress: the bare world press (the Interact fallthrough,
	// plrctrls.cpp:2345).
	KindPrimaryActionPress Kind = "PrimaryActionPress"
	// KindSpellSelect: the speed-book selection press assigns the spell to the
	// slot (plrctrls.cpp:2311-2314 SetSpell).
	KindSpellSelect Kind = "SpellSelect"
	// KindIconCursorAct: an item-targeting cursor acted on its target
	// (plrctrls.cpp:2318-2322 TryIconCurs + NewCursor(CURSOR_HAND)); A carries
	// the pcurs ordinal acted on.
	KindIconCursorAct Kind = "IconCursorAct"
	// KindInventoryLift: lift the inventory item under the pointer
	// (plrctrls.cpp:2323-2324 LiftInventoryItem); A carries the resolved
	// controls.PanelRegion.
	KindInventoryLift Kind = "InventoryLift"
	// KindStashLift: lift the stash item under the pointer
	// (plrctrls.cpp:2325-2326 LiftStashItem).
	KindStashLift Kind = "StashLift"
	// KindStatAllocate: the char-panel stat allocation press
	// (plrctrls.cpp:2340-2344 CheckChrBtns/ReleaseChrBtns).
	KindStatAllocate Kind = "StatAllocate"
	// KindStoreAct: one visual-store action; A is the verb
	// (inputbridge StoreAct* values), B the pcursstorebtn ordinal for the
	// button verb (plrctrls.cpp:2328-2336, :2350-2352).
	KindStoreAct Kind = "StoreAct"
	// KindHourglassNoop: the hourglass-cursor press — upstream returns without
	// firing an action and without resetting the cursor (plrctrls.cpp:2319-2320);
	// a handler-free kind that keeps the refusal in the log.
	KindHourglassNoop Kind = "HourglassNoop"
	// KindInventorySwallow: the inventory-open click that falls through the
	// invflag block's branches — upstream returns having fired nothing
	// (plrctrls.cpp:2337); a handler-free kind like KindHourglassNoop.
	KindInventorySwallow Kind = "InventorySwallow"

	// Input bridge, mouse-path family — the LeftMouseCmd rows (diablo.cpp
	// 247-319 @ fork ca91d778f, devilGo-c6wt s2). The target tile travels in
	// the mouse-path command payload words; the kind names the verb row.
	// KindMouseWalk: the walk fallthrough (diablo.cpp:261-263 town,
	// :317-319 dungeon); A=x, B=y of cursPosition.
	KindMouseWalk Kind = "MouseWalk"
	// KindMouseTalk: the town towner talk (diablo.cpp:258-259 CMD_TALKXY);
	// A is pcursmonst.
	KindMouseTalk Kind = "MouseTalk"
	// KindMousePickup: the item pickup (diablo.cpp:255-256 CMD_GOTOGETITEM /
	// CMD_GOTOAGETITEM); A is pcursitem, B the near flag (invflag).
	KindMousePickup Kind = "MousePickup"
	// KindMouseObject: the object operate/disarm (diablo.cpp:273-275
	// CMD_DISARMXY / CMD_OPOBJXY); A the object id, B the disarm flag.
	KindMouseObject Kind = "MouseObject"
	// KindMouseAttack: the dungeon attack family (diablo.cpp:277-316); A the
	// ranged flag (weapon truth: RATTACK* over SATTACK*/ATTACK*), B the verb
	// word (0 = plain id attack, 1 = tile attack, 2 = talk — the
	// CanTalkToMonst ATTACKID, reachable under a ranged weapon per
	// diablo.cpp:283-284), C the id target (monster id; a player id shares
	// the word per ADR-015 integer payload — the handler disambiguates by
	// the target registry).
	KindMouseAttack Kind = "MouseAttack"

	// Session verbs — the transitions the level and the floor own, as opposed
	// to the input rows above. They carry no tile: the verb acts on where the
	// player already stands.
	// KindDescend loads a dungeon level (LoadGameLevel, diablo.cpp:3142-3200);
	// A carries the level number, B the dungeon_type the caller expects.
	KindDescend Kind = "Descend"
	// KindLoot takes the item lying on the player's tile (PickItem,
	// inv.cpp:1694-1702); no payload — the tile is the player's own.
	KindLoot Kind = "Loot"
	// KindDescended announces the level load: A the level number, B the
	// dungeon_type that landed, C the tile the player arrived on, packed as
	// Y*MAXDUNX+X (the packing the MonsterDied payload already uses).
	KindDescended Kind = "Descended"
	// KindLooted announces the pickup: A the Items index taken off the floor,
	// B and C the tile it was lying on (CleanupItems, inv.cpp:1053-1057).
	KindLooted Kind = "Looted"
	// KindCreatePack raises the session's hero from its class's starting pack
	// (CreatePlayer, player.cpp:2283-2330); A is the class ordinal, B and C the
	// spawn tile the hero is raised on.
	KindCreatePack Kind = "CreatePack"
	// KindCreatedPack announces the raised hero: A the class ordinal the pack
	// belongs to, B and C the tile the hero now stands on.
	KindCreatedPack Kind = "CreatedPack"
)

// Kind identifies a command or an event type by domain name. The literal
// string is part of the trace hash contract, so renaming a kind changes hashes
// (ADR-018 law 5: names live in one package).
type Kind string

// Command is one intent applied to the simulation. The payload is integer-only
// per ADR-015: no float64 enters a per-tick path.
type Command struct {
	// Tick is the tick the command belongs to (ADR-018 law 1); dispatch stamps
	// events from the pass tick, so this value documents intent only.
	Tick uint32
	// Kind names the command.
	Kind Kind
	// A, B and C are integer payload words, interpreted per Kind.
	A, B, C int32
	// Down is the pressed state for input commands.
	Down bool
}

// Event is one applied state change: the unit presentation reads and the tick
// log records (ADR-018 laws 2 and 4).
type Event struct {
	// Tick is the tick whose pass emitted the event.
	Tick uint32
	// Kind names the event.
	Kind Kind
	// A, B and C mirror the originating command payload words.
	A, B, C int32
	// Stream is the RNG stream the emitting handler drew from, empty when it
	// drew nothing.
	Stream string
	// Roll is the value drawn from Stream, when one was drawn.
	Roll uint32
}

// Gate decides whether a command may be dispatched. A dropped command emits no
// event (ADR-018 law 6).
type Gate interface {
	Allow(Command) bool
}

// TickRecord is one finished pass: its dispatched commands (the replay
// input), the events it emitted and the events' checksum — the row the tick
// log stores and replays (ADR-018 law 4). persist.TickRecord mirrors this
// shape at the store boundary.
type TickRecord struct {
	// Tick is the pass tick.
	Tick uint32
	// Commands are the commands dispatched in that pass, in submission order.
	Commands []Command
	// Events are the events the pass emitted, in trace order.
	Events []Event
	// Checksum is TraceHash over this tick's events.
	Checksum uint64
}

// Handler applies one command and returns the events it caused. An empty result
// is the honest answer for a command that changed nothing.
type Handler func(Command, *Streams) []Event
