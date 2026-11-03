package druid

// The November 2 development notes (Blizzard, 2026-11-02, the notes for client 1.60.1.70753) raise
// the cat's and Balance's damage:
//
//   - Shred 200% weapon damage (was 155%), Claw 140% (110%), Ravage 435% (350%): shred.go,
//     claw.go, ravage.go.
//   - "Rip / Rake / Ferocious Bite / Pounce: Base points and attack power scaling increased by
//     about 25%": nov2BleedScale on the rank tables and the attack power coefficients.
//   - "Starfire: Base damage points increased by about 40%": nov2StarfireScale on the base damage
//     and its growth a level.
//   - Rend and Tear 3% a point (2% before): talents.go.
//
// None of it is in the client: 1.60.1.70753's SpellEffect rows for these spells are those of
// 70707, Wowhead (dataEnv 16) shows the old values, and wago.tools has no hotfix rows for them
// (checked 2026-11-03). It's server tuning the client doesn't carry yet, so the sim takes the notes,
// with "about" read as the round number (the report's single-value assumption; see
// orchestration/round_70753/druid.md).
const (
	nov2BleedScale    = 1.25
	nov2StarfireScale = 1.4
)
