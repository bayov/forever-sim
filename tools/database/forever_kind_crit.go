package database

// KindLimitedCrit says whether the item's crit counts for one kind of attack only.
//
// Forever's gear crit counts for melee, ranged and spells alike, and its tooltip reads
// "Improves your chance to get a critical strike by 1.0%". Wowhead lists that as crit
// rating (critstrkrtng). A few items keep a tooltip limited to one kind, which Wowhead
// lists as a percent of melee and ranged (mlecritstrkpct, rgdcritstrkpct) or of spells
// (splcritstrkpct). The game confirms both we checked: Blackhand's Breadth reads "with
// Melee and Ranged attacks" and Eye of the Beast "with Spells".
func (wi ForeverWowheadItem) KindLimitedCrit() bool {
	return wi.num("mlecritstrkpct") != 0 || wi.num("rgdcritstrkpct") != 0 || wi.num("splcritstrkpct") != 0
}
