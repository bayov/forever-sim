package paladin

// Light's Vigil (the 31 point Holy talent, needs Holy Shock) heals only since the launch realms'
// October 30 hotfix (spell 1310911 in 1.60.1.70707+hf, wago.tools hotfixes region 91): "Applies
// Light's Vigil to a friendly target for 30 sec. The caster's next Holy Shock on the target incurs
// no cooldown and heals the target and their party." Its effects now target an ally (implicit
// target 21, was 25, any unit) and the enemy damage and the 75% mana refund are out of the
// tooltip. A Holy Shock on the enemy no longer touches it, so the sim, which does not model the
// paladin's heals, registers nothing for it. Before the hotfix the Vigil went on the enemy, and the
// next Holy Shock on it skipped its cooldown and detonated it for Holy damage (the Holy Shock
// builds were built on that).
var LightsVigilLevel = [...]int{0, 40, 50, 60}
var LightsVigilSpellId = [...]int32{0, 1310911, 1311590, 1311595}

// LightsVigilDebuffId is the mark each rank puts on its target.
var LightsVigilDebuffId = [...]int32{0, 1310909, 1311593, 1311597}

func (paladin *Paladin) registerLightsVigil() {}
