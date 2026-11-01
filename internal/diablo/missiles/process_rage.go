// Package missiles: the barbarian Rage skill's bodies.
package missiles

// Spell flag bits AddRage/ProcessRage toggle (player.h:152-153).
const (
	// SpellFlagRageActive is SpellFlag::RageActive.
	SpellFlagRageActive = 1 << 1
	// SpellFlagRageCooldown is SpellFlag::RageCooldown.
	SpellFlagRageCooldown = 1 << 2
)

// Hero speech ordinals the rage bodies play (sound_effect_enums.h:79, 81).
const (
	// SpeechAaaaargh is HeroSpeech::Aaaaargh.
	SpeechAaaaargh = 79
	// SpeechHeavyBreathing is HeroSpeech::HeavyBreathing.
	SpeechHeavyBreathing = 81
)

// AddRage ports AddRage (missiles.cpp:2879-2897): rage refuses to stack, so an
// armed or cooling-down caster — or one whose pooled life is no bigger than its
// level — fizzles (cpp:2882-2885). Otherwise the damage is level * 6 and the
// life 245 + level * 2 (cpp:2886-2888), and the active flag goes up before the
// item values are recomputed (cpp:2892-2894).
func AddRage(w *HitWorld, m *Missile, p *AddParameter) {
	stats := w.AddSource.playerStats(m.Source)
	flags := w.ProcessCallbacks.PlayerSpellFlags(m.Source)

	if flags&(SpellFlagRageActive|SpellFlagRageCooldown) != 0 || // cpp:2882-2884
		stats.HitPoints <= stats.Level<<6 {
		m.DelFlag = true
		p.SpellFizzled = true

		return
	}

	m.Dam = stats.Level * 6          // cpp:2886
	m.Duration = 245 + stats.Level*2 // cpp:2887
	m.Var1 = m.Duration              // cpp:2888

	w.ProcessCallbacks.SetPlayerSpellFlags(m.Source, flags|SpellFlagRageActive) // cpp:2890
	w.ProcessCallbacks.RecalcPlayerItems(m.Source)                              // cpp:2891
	w.ProcessCallbacks.Say(m.Source, SpeechAaaaargh)                            // cpp:2892
}

// ProcessRage ports ProcessRage (missiles.cpp:3625-3660): the first expiry
// hands the active flag over to the cooldown and restarts the timer
// (cpp:3632-3638), the second expiry clears the cooldown and drops the missile
// (cpp:3639-3641). Both expiries recompute the item values, floor the life at
// one point so the recompute cannot kill the caster (cpp:3647-3652) and redraw.
func ProcessRage(w *HitWorld, m *Missile) {
	m.Duration-- // cpp:3627

	if m.Duration != 0 { // cpp:3629-3630
		return
	}

	flags := w.ProcessCallbacks.PlayerSpellFlags(m.Source) // cpp:3632

	switch {
	case flags&SpellFlagRageActive != 0: // cpp:3633-3638
		flags = flags&^SpellFlagRageActive | SpellFlagRageCooldown
		m.Duration = m.Var1
	case flags&SpellFlagRageCooldown != 0: // cpp:3639-3641
		flags &^= SpellFlagRageCooldown
		m.DelFlag = true
	default: // neither flag is up: upstream leaves the flags and the timer alone
	}

	w.ProcessCallbacks.SetPlayerSpellFlags(m.Source, flags)
	w.ProcessCallbacks.RecalcPlayerItems(m.Source) // cpp:3645

	if w.AddSource.playerStats(m.Source).HitPoints <= 0 { // cpp:3647-3652
		w.ProcessCallbacks.SetPlayerHitPoints(m.Source, fixed26_6One)
	}

	w.ProcessCallbacks.RedrawEverything()                  // cpp:3653
	w.ProcessCallbacks.Say(m.Source, SpeechHeavyBreathing) // cpp:3654

	if m.DelFlag { // cpp:3656-3659
		w.ProcessCallbacks.DamagePlayer(m.Source, int(DamagePhysical), m.Dam<<6, 1)
	}
}
