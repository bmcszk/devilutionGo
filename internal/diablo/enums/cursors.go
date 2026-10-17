// Package enums centralizes cross-cutting constants that are not TSV-driven
// (devilGo-orkj; upstream sources cited per declaration).
package enums

// SelectionRegion mirrors cursor_defs.hpp:14-19 (bit flags).
type SelectionRegion uint8

const (
	SelectionNone   SelectionRegion = 0
	SelectionBottom SelectionRegion = 1 << 0
	SelectionMiddle SelectionRegion = 1 << 1
	SelectionTop    SelectionRegion = 1 << 2
)

// CursorID mirrors cursor_id (cursor_defs.hpp:22-36), all 13 verbatim.
type CursorID uint8

const (
	CursorNone        CursorID = iota // CURSOR_NONE
	CursorHand                        // CURSOR_HAND
	CursorIdentify                    // CURSOR_IDENTIFY
	CursorRepair                      // CURSOR_REPAIR
	CursorRecharge                    // CURSOR_RECHARGE
	CursorDisarm                      // CURSOR_DISARM
	CursorOil                         // CURSOR_OIL
	CursorTelekinesis                 // CURSOR_TELEKINESIS
	CursorResurrect                   // CURSOR_RESURRECT
	CursorTeleport                    // CURSOR_TELEPORT
	CursorHealOther                   // CURSOR_HEALOTHER
	CursorHourglass                   // CURSOR_HOURGLASS
	CursorFirstItem                   // CURSOR_FIRSTITEM
)
