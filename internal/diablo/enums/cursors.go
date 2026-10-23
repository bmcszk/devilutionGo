// Package enums centralizes cross-cutting constants that are not TSV-driven.
package enums

// SelectionRegion is a bit flag set of selectable regions.
type SelectionRegion uint8

const (
	SelectionNone   SelectionRegion = 0
	SelectionBottom SelectionRegion = 1 << 0
	SelectionMiddle SelectionRegion = 1 << 1
	SelectionTop    SelectionRegion = 1 << 2
)

// cursorID identifies a mouse cursor; the constant order is the 1:1 contract.
type cursorID uint8

const (
	CursorNone cursorID = iota
	CursorHand
	CursorIdentify
	CursorRepair
	CursorRecharge
	CursorDisarm
	CursorOil
	CursorTelekinesis
	CursorResurrect
	CursorTeleport
	CursorHealOther
	CursorHourglass
	CursorFirstItem
)
