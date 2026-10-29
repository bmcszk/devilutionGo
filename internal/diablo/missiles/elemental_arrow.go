// Elemental-arrow explosion port (missiles.cpp:2848-2930): the trap dice, the
// explosion row switch and the flight tick that lands it.
package missiles

import (
	"fmt"

	"github.com/bmcszk/devilutionGo/internal/engine/appfat"
)

// elementalPack bundles the explosion sheet, element and damage dice of one
// elemental arrow row (language-artifact wrapper; upstream returns 4 locals).
type elementalPack struct {
	graphic        GraphicID
	damageType     DamageType
	minDam, maxDam int
}

func elementalTrapDamage(w *HitWorld) (minDamage, maxDamage int) {
	return w.ProcessCallbacks.GenerateRnd(10) + 1 + w.ProcessCallbacks.Currlevel,
		w.ProcessCallbacks.GenerateRnd(10) + 1 + w.ProcessCallbacks.Currlevel*2
}

// elementalExplosionPack resolves the explosion row of an elemental arrow
// (missiles.cpp:2913-2921); a non-elemental type is the upstream app_fatal
// "wrong missile ID" trip, returned instead of terminating.
func elementalExplosionPack(w *HitWorld, m *Missile) (elementalPack, *appfat.FatalError) {
	switch m.Type {
	case LightningArrow:
		return elementalPackFor(w, m, GraphicChargedBolt, DamageLightning), nil
	case FireArrow:
		return elementalPackFor(w, m, GraphicMagmaBallExplosion, DamageFire), nil
	default:
		return elementalPack{}, appfat.AppFatal(fmt.Sprintf("wrong missile ID %d", int(m.Type)))
	}
}

func elementalPackFor(w *HitWorld, m *Missile, eAnim GraphicID, damageType DamageType) elementalPack {
	if m.IsTrap() {
		eMind, eMaxd := elementalTrapDamage(w)
		return elementalPack{graphic: eAnim, damageType: damageType, minDam: eMind, maxDam: eMaxd}
	}
	eMind, eMaxd := w.ProcessCallbacks.ElementDamage(m, damageType)
	return elementalPack{graphic: eAnim, damageType: damageType, minDam: eMind, maxDam: eMaxd}
}

func processElementalArrowFlight(w *HitWorld, m *Missile) {
	m.Dist++
	var mind, maxd int
	if m.IsTrap() {
		mind, maxd = elementalTrapDamage(w)
	} else {
		mind, maxd = w.ProcessCallbacks.SourceDamage(m)
	}
	MoveMissileAndCheckMissileCol(w, m, DamagePhysical, mind, maxd, true, false)

	if m.Duration == 0 {
		// Explosion transition.
		m.setDefaultFrameGroup()
		m.Duration = m.AnimLen - 1
		StopMissile(m)
		pack, fatal := elementalExplosionPack(w, m)
		if fatal != nil {
			w.ProcessCallbacks.Fatal(fatal)

			return
		}

		m.setAnimation(pack.graphic)
		// The damage re-probe passes dontDeleteOnCollision = true
		// keepOnCollision.
		CheckMissileCol(w, m, m.Position.Tile, MissileColProbe{
			DamageType:     pack.damageType,
			MinDamage:      pack.minDam,
			MaxDamage:      pack.maxDam,
			Shifted:        false,
			Keep:           true,
			OnlyHitWalking: nil,
		})
		return
	}
	if m.Position.Tile != (TilePosition{X: m.Var1, Y: m.Var2}) {
		// Flight tile tracker: radius 5.
		m.Var1 = m.Position.Tile.X
		m.Var2 = m.Position.Tile.Y
		w.ProcessCallbacks.ChangeLight(m.Lid, m.Position.Tile.X, m.Position.Tile.Y, 5)
	}
}
