// Package stores ports the store core from stores.cpp at devilutionX
// c20b2f2456: the TalkID dialog state machine, the transaction price rules,
// the premium stock refresh and the towner gossip table. It is Diablo-specific
// game logic (AGENTS.md layer 3), so it lives next to the other diablo
// packages; internal/gameplay/stores re-exports the surface the UI epic
// imports, the same split internal/diablo/inventory uses.
//
// Rendering, input, sound and network sends are dropped: they are game-loop
// side effects, not store state. Every ported branch cites its upstream line.
package stores

// TalkID is the dialog page currently shown (TalkID, stores.h:44-69). The
// ordinal is upstream's declaration order, which saves encode and tests pin.
type TalkID uint8

// The dialog pages, in upstream declaration order (stores.h:44-69).
const (
	TalkIDNone TalkID = iota
	TalkIDSmith
	TalkIDSmithBuy
	TalkIDSmithSell
	TalkIDSmithRepair
	TalkIDWitch
	TalkIDWitchBuy
	TalkIDWitchSell
	TalkIDWitchRecharge
	TalkIDNoMoney
	TalkIDNoRoom
	TalkIDConfirm
	TalkIDBoy
	TalkIDBoyBuy
	TalkIDHealer
	TalkIDStoryteller
	TalkIDHealerBuy
	TalkIDStorytellerIdentify
	TalkIDSmithPremiumBuy
	TalkIDGossip
	TalkIDStorytellerIdentifyShow
	TalkIDTavern
	TalkIDDrunk
	TalkIDBarmaid
)

// TalkIDCount is the number of TalkID states, i.e. the value one past
// TalkIDBarmaid (stores.h:44-69).
const TalkIDCount = int(TalkIDBarmaid) + 1

var talkIDNames = [TalkIDCount]string{
	"None", "Smith", "SmithBuy", "SmithSell", "SmithRepair",
	"Witch", "WitchBuy", "WitchSell", "WitchRecharge",
	"NoMoney", "NoRoom", "Confirm",
	"Boy", "BoyBuy",
	"Healer", "Storyteller", "HealerBuy", "StorytellerIdentify",
	"SmithPremiumBuy", "Gossip", "StorytellerIdentifyShow",
	"Tavern", "Drunk", "Barmaid",
}

// String returns upstream's enumerator name, or "TalkID(n)" when the value is
// out of range.
func (t TalkID) String() string {
	if int(t) < TalkIDCount {
		return talkIDNames[t]
	}

	return "TalkID(" + itoa(int(t)) + ")"
}

// townerShortNames maps a top-level store page to the towner short name the
// Lua StoreOpened event and mod dialog options key on (stores.cpp:78-92).
var townerShortNames = map[TalkID]string{
	TalkIDSmith:       "griswold",
	TalkIDWitch:       "adria",
	TalkIDBoy:         "wirt",
	TalkIDHealer:      "pepin",
	TalkIDStoryteller: "cain",
	TalkIDTavern:      "ogden",
	TalkIDDrunk:       "farnham",
	TalkIDBarmaid:     "gillian",
}

// TownerNameForTalkID returns the towner short name for a top-level store
// page, and false for a sub-page such as TalkIDSmithBuy (stores.cpp:78-92).
func TownerNameForTalkID(t TalkID) (string, bool) {
	name, ok := townerShortNames[t]

	return name, ok
}

// itoa is strconv.Itoa for the one non-table path in String, kept local so the
// package needs no import for it.
func itoa(v int) string {
	if v == 0 {
		return "0"
	}

	var digits []byte
	for v > 0 {
		digits = append([]byte{byte('0' + v%10)}, digits...)
		v /= 10
	}

	return string(digits)
}
