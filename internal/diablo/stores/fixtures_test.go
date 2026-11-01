package stores_test

import (
	"github.com/bmcszk/devilutionGo/internal/diablo/entities"
	"github.com/bmcszk/devilutionGo/internal/diablo/inventory"
)

// The test fixtures build players the way the game loop does: every slot seeded
// with inventory.EmptyItem, because C++ default-constructs Item with
// _itype = ItemType::None and the Go zero value cannot express it.

// playerFixture wraps the player so the put* helpers stay terse.
type playerFixture struct {
	player *entities.Player
}

// newPlayer returns a player with an empty body, inventory and belt.
func newPlayer() *entities.Player {
	player := &entities.Player{}

	for i := range player.InvBody {
		player.InvBody[i] = inventory.EmptyItem()
	}

	for i := range player.InvList {
		player.InvList[i] = inventory.EmptyItem()
	}

	for i := range player.SpdList {
		player.SpdList[i] = inventory.EmptyItem()
	}

	return player
}

// putInv places an item in inventory cell i and grows _pNumInv past it.
func (p *playerFixture) putInv(i int, item entities.Item) {
	p.player.InvList[i] = item
	if i+1 > p.player.NumInv {
		p.player.NumInv = i + 1
	}
}

// putBelt places an item in belt slot i.
func (p *playerFixture) putBelt(i int, item entities.Item) {
	p.player.SpdList[i] = item
}

// putBody places an item in an equipment slot (player.h:73-82).
func (p *playerFixture) putBody(loc inventory.BodyLoc, item entities.Item) {
	p.player.InvBody[loc] = item
}

// fixture returns a player fixture.
func fixture() *playerFixture {
	return &playerFixture{player: newPlayer()}
}

// sword is a plain, usable weapon worth 1000.
func sword() entities.Item {
	return entities.Item{
		Itype:      int8(inventory.ItemTypeSword),
		IDidx:      1,
		Value:      1000,
		Ivalue:     4000,
		MaxDur:     100,
		Durability: 100,
		StatFlag:   true,
	}
}

// damagedSword is sword() at half durability.
func damagedSword() entities.Item {
	item := sword()
	item.Durability = 50

	return item
}
