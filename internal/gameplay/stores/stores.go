// Package stores re-exports the store core the bean names for this tree. The
// port itself lives in internal/diablo/stores (AGENTS.md layer 3: Diablo store
// and towner logic), so this package is type aliases and function values only —
// the same split internal/gameplay already uses for inventory and itemgen.
//
// Layering decision (recorded, per the bean contract): the contract asked for
// internal/gameplay/stores; the existing tree puts every Diablo-specific port
// under internal/diablo/… and keeps internal/gameplay as the game-facing
// facade, so the logic went to internal/diablo/stores and this package stays
// the facade the UI epic imports.
package stores

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/stores"
)

// The TalkID pages, in upstream order.
const (
	TalkIDNone                    = stores.TalkIDNone
	TalkIDSmith                   = stores.TalkIDSmith
	TalkIDSmithBuy                = stores.TalkIDSmithBuy
	TalkIDSmithSell               = stores.TalkIDSmithSell
	TalkIDSmithRepair             = stores.TalkIDSmithRepair
	TalkIDWitch                   = stores.TalkIDWitch
	TalkIDWitchBuy                = stores.TalkIDWitchBuy
	TalkIDWitchSell               = stores.TalkIDWitchSell
	TalkIDWitchRecharge           = stores.TalkIDWitchRecharge
	TalkIDNoMoney                 = stores.TalkIDNoMoney
	TalkIDNoRoom                  = stores.TalkIDNoRoom
	TalkIDConfirm                 = stores.TalkIDConfirm
	TalkIDBoy                     = stores.TalkIDBoy
	TalkIDBoyBuy                  = stores.TalkIDBoyBuy
	TalkIDHealer                  = stores.TalkIDHealer
	TalkIDStoryteller             = stores.TalkIDStoryteller
	TalkIDHealerBuy               = stores.TalkIDHealerBuy
	TalkIDStorytellerIdentify     = stores.TalkIDStorytellerIdentify
	TalkIDSmithPremiumBuy         = stores.TalkIDSmithPremiumBuy
	TalkIDGossip                  = stores.TalkIDGossip
	TalkIDStorytellerIdentifyShow = stores.TalkIDStorytellerIdentifyShow
	TalkIDTavern                  = stores.TalkIDTavern
	TalkIDDrunk                   = stores.TalkIDDrunk
	TalkIDBarmaid                 = stores.TalkIDBarmaid
	// TalkIDCount is the page count.
	TalkIDCount = stores.TalkIDCount
)

// The panel layout and transaction constants.
const (
	// NumStoreLines is the panel line count.
	NumStoreLines = stores.NumStoreLines
	// BackButtonLine is the menu back button line.
	BackButtonLine = stores.BackButtonLine
	// IdentifyPrice is Cain's flat identify fee.
	IdentifyPrice = stores.IdentifyPrice
	// MaxHoldItems is the offered-list capacity.
	MaxHoldItems = stores.MaxHoldItems
	// NumSmithItems and NumSmithItemsHF are the premium shelf sizes.
	NumSmithItems   = stores.NumSmithItems
	NumSmithItemsHF = stores.NumSmithItemsHF
)

// The line kinds, alignments, colors and actions.
const (
	LineLabel      = stores.LineLabel
	LineDivider    = stores.LineDivider
	LineSelectable = stores.LineSelectable

	AlignDefault = stores.AlignDefault
	AlignCenter  = stores.AlignCenter
	AlignRight   = stores.AlignRight

	ColorDefault   = stores.ColorDefault
	ColorWhite     = stores.ColorWhite
	ColorBlue      = stores.ColorBlue
	ColorWhitegold = stores.ColorWhitegold
	ColorRed       = stores.ColorRed

	ActionNone            = stores.ActionNone
	ActionBuyBasic        = stores.ActionBuyBasic
	ActionBuyPremium      = stores.ActionBuyPremium
	ActionSell            = stores.ActionSell
	ActionRepair          = stores.ActionRepair
	ActionBuyWitch        = stores.ActionBuyWitch
	ActionRecharge        = stores.ActionRecharge
	ActionIdentify        = stores.ActionIdentify
	ActionBuyBoy          = stores.ActionBuyBoy
	ActionBuyHealer       = stores.ActionBuyHealer
	ActionOpenVisualStore = stores.ActionOpenVisualStore
	ActionOpenStorage     = stores.ActionOpenStorage
	ActionGossipText      = stores.ActionGossipText
	StoreUIList           = stores.StoreUIList
	StoreUIVisualGrid     = stores.StoreUIVisualGrid
	HoldHead              = stores.HoldHead
	HoldChest             = stores.HoldChest
	HoldHandLeft          = stores.HoldHandLeft
	HoldHandRight         = stores.HoldHandRight
	HoldRingLeft          = stores.HoldRingLeft
	HoldRingRight         = stores.HoldRingRight
	HoldAmulet            = stores.HoldAmulet
)

// ErrUnknownStore is the invalid-page error.
var ErrUnknownStore = stores.ErrUnknownStore

// GossipTextIDs is the towner gossip id table.
var GossipTextIDs = stores.GossipTextIDs

// The entry points, re-exported from internal/diablo/stores.
var (
	// TownerNameForTalkID maps a top-level page to its towner short name.
	TownerNameForTalkID = stores.TownerNameForTalkID
	// PickGossip draws the gossip text id for a towner.
	PickGossip = stores.PickGossip
	// GossipText resolves a gossip id through the textdat table.
	GossipText = stores.GossipText
	// SellValue is the quarter-value sell price.
	SellValue = stores.SellValue
	// RepairPrice is the durability repair price.
	RepairPrice = stores.RepairPrice
	// RechargePrice is the staff recharge price.
	RechargePrice = stores.RechargePrice
	// BoyPrice is Wirt's haggling price.
	BoyPrice = stores.BoyPrice
	// TotalGold is carried plus stashed gold.
	TotalGold = stores.TotalGold
	// CanAfford is the gold check.
	CanAfford = stores.CanAfford
	// GoldFit is StoreGoldFit.
	GoldFit = stores.GoldFit
	// SmithWillBuy is Griswold's buy rule.
	SmithWillBuy = stores.SmithWillBuy
	// WitchWillBuy is Adria's buy rule.
	WitchWillBuy = stores.WitchWillBuy
	// Repairable is SmithRepairOk.
	Repairable = stores.Repairable
	// Rechargeable is WitchRechargeOk.
	Rechargeable = stores.Rechargeable
	// Identifiable is IdItemOk.
	Identifiable = stores.Identifiable
	// TextColor is getTextColorWithStatCheck.
	TextColor = stores.TextColor
	// BodySlot maps a hold id to its body slot.
	BodySlot = stores.BodySlot
	// SmithSellHolds builds Griswold's buy-from-player list.
	SmithSellHolds = stores.SmithSellHolds
	// WitchSellHolds builds Adria's buy-from-player list.
	WitchSellHolds = stores.WitchSellHolds
	// RepairHolds builds the repair list.
	RepairHolds = stores.RepairHolds
	// RechargeHolds builds the recharge list.
	RechargeHolds = stores.RechargeHolds
	// IdentifyHolds builds Cain's identify list.
	IdentifyHolds = stores.IdentifyHolds
	// SpawnPremium fills and refreshes the premium shelf.
	SpawnPremium = stores.SpawnPremium[entities.Item]
	// ReplacePremium refills one premium slot.
	ReplacePremium = stores.ReplacePremium[entities.Item]
	// ResetPremium clears the premium shelf the way InitStores does.
	ResetPremium = stores.ResetPremium[entities.Item]
	// RefreshPremium is SetupTownStores' premium step on town entry.
	RefreshPremium = stores.RefreshPremium[entities.Item]
	// StoreLevel is SetupTownStores' shared vendor level rule.
	StoreLevel = stores.StoreLevel
)

// TalkID, Line, Dialog and Session are the panel state the UI epic consumes.
type (
	// TalkID is the active dialog page.
	TalkID = stores.TalkID
	// Line is one panel line.
	Line = stores.Line
	// Dialog is the panel: 104 lines plus the scroll flags.
	Dialog = stores.Dialog
	// Session is the store dialog state machine.
	Session = stores.Session
	// GameView is the world state the pages read.
	GameView = stores.GameView
	// MenuInputs is the store-UI mode and talk inputs.
	MenuInputs = stores.MenuInputs
	// Hold is one offered item.
	Hold = stores.Hold
	// Outcome is what one Enter produced.
	Outcome = stores.Outcome
	// Action is the executed trade kind.
	Action = stores.Action
	// Env is the game-mode branch set.
	Env = stores.Env
	// StoreUI is the StoreUi option value.
	StoreUI = stores.StoreUI
	// PremiumStockItem is Griswold's rotating shelf of items.
	PremiumStockItem = stores.PremiumStock[entities.Item]
)
