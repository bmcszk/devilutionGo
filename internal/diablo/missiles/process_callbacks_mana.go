// Package missiles: the mana and reflection sinks of ProcessCallbacks, split
// out to keep process.go within the file-length limit.
package missiles

// manaHooks groups the mana and reflection sinks (player.h:236-237,
// missiles.cpp:1278-1297, 1458-1476, 1655-1684). ProcessCallbacks embeds it, so
// the fields stay reachable as cb.PlayerMana and friends.
type manaHooks struct {
	// PlayerMana reads the mana pool of one player (player.h:236-237,
	// missiles.cpp:1657).
	PlayerMana func(player int) PlayerMana
	// SetPlayerMana writes _pMana and _pManaBase (missiles.cpp:1666-1669).
	SetPlayerMana func(player, mana, manaBase int)
	// ClassManaRestore reads GetClassAttributes(class).splRestoreMana, a
	// Fixed26_6 (missiles.cpp:1670-1671).
	ClassManaRestore func(class int) int
	// RedrawMana ports RedrawComponent(PanelDrawComponent::Mana)
	// (missiles.cpp:1673).
	RedrawMana func()
	// CalcPlayerInv ports CalcPlrInv(player, false) (items.cpp:2903,
	// missiles.cpp:1470).
	CalcPlayerInv func(player int)
	// PlayerReflections reads Player::wReflections (player.h,
	// missiles.cpp:1292).
	PlayerReflections func(player int) int
	// SetPlayerReflections writes Player::wReflections (missiles.cpp:1294).
	SetPlayerReflections func(player, reflections int)
	// SendSetReflect ports NetSendCmdParam1(CMD_SETREFLECT, wReflections)
	// (missiles.cpp:1296).
	SendSetReflect func(reflections int)
}
