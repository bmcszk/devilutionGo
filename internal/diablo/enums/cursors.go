// Package enums centralizes cross-cutting constants that are not TSV-driven.
package enums

const (
	SelectionNone   SelectionRegion = 0
	SelectionBottom SelectionRegion = 1 << 0
	SelectionMiddle SelectionRegion = 1 << 1
	SelectionTop    SelectionRegion = 1 << 2
)

const (
	CursorNone CursorID = iota
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

// SelectionRegion is a bit flag set of selectable regions.
type SelectionRegion uint8

// CursorID identifies a mouse cursor; the constant order is the 1:1 contract.
type CursorID uint8
