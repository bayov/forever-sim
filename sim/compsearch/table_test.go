package compsearch

import (
	"slices"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// The buff sets' toggles come from the table's specs and duties. They're what the search had by hand
// before the table, except that Moonkin Aura and Leader of the Pack now bring Gift of the Wild: every
// build that gives them is a druid, which gives it too.
func TestToggles(t *testing.T) {
	want := []Gift{GiftFort | GiftPoS, GiftGotW, GiftAB, GiftBattleShout, GiftFuriousHowl, GiftTrueshot, GiftLotP | GiftGotW,
		GiftMoonkin | GiftGotW, GiftSoE | GiftWF, GiftSunderTank, GiftMinorArmor, GiftCoE, GiftCrusader, GiftJoW}
	if !slices.Equal(toggles, want) {
		var got []string
		for _, g := range toggles {
			got = append(got, g.String())
		}
		t.Errorf("toggles %q", got)
	}
}

func TestBlessingChoices(t *testing.T) {
	k, m, w := proto.Blessing_BlessingKings, proto.Blessing_BlessingMight, proto.Blessing_BlessingWisdom
	for n, want := range map[int][][]proto.Blessing{
		0: {nil},
		1: {{k}, {m}, {w}},
		2: {{k, m}, {k, w}, {m, w}},
		3: {{k, m, w}},
	} {
		got := blessingChoices(n)
		if !slices.EqualFunc(got, want, func(a, b []proto.Blessing) bool { return slices.Equal(a, b) }) {
			t.Errorf("%d paladins: %v, want %v", n, got, want)
		}
	}
	if maxPaladins != 3 {
		t.Errorf("maxPaladins %d", maxPaladins)
	}
}

func TestTableNames(t *testing.T) {
	if g, ok := GiftByName("Salvation"); !ok || g != GiftBlessing {
		t.Errorf("Salvation is %v", g)
	}
	if GiftName(GiftWF) != "Windfury Totem" || totemName(proto.RaidTotem_RaidTotemGraceOfAir) != "Grace of Air" {
		t.Errorf("totem names")
	}
	if !slices.Equal(fixedTotems, []proto.RaidTotem{proto.RaidTotem_RaidTotemStrengthOfEarth, proto.RaidTotem_RaidTotemManaSpring,
		proto.RaidTotem_RaidTotemFlametongue}) {
		t.Errorf("fixed totems %v", fixedTotems)
	}
	if fixedTotemGifts != GiftSoE|GiftFT || otherAir(proto.RaidTotem_RaidTotemWindfury) != proto.RaidTotem_RaidTotemGraceOfAir {
		t.Errorf("shaman placeholder totems")
	}
	for _, p := range Presets {
		if p.Paladin() != (p.Class == proto.Class_ClassPaladin) || p.Shaman() != (p.Class == proto.Class_ClassShaman) {
			t.Errorf("%s: paladin %v, shaman %v", p.Name, p.Paladin(), p.Shaman())
		}
	}
}

// Every DPS build of the pool has a spec row, and the rows say what the class checks said before them.
func TestSpecRows(t *testing.T) {
	p := loadTestPool(t)
	for _, b := range p.Builds {
		s := b.SpecRow()
		if s == noSpec {
			t.Errorf("%s has no spec row", b.ID)
			continue
		}
		feral := b.Spec == "Feral"
		if s.Mana != (b.Class != proto.Class_ClassWarrior && b.Class != proto.Class_ClassRogue && !feral) {
			t.Errorf("%s: mana %v", b.ID, s.Mana)
		}
		if blesses(b) != (b.Class == proto.Class_ClassPaladin) || dropsTotems(b) != (b.Class == proto.Class_ClassShaman) {
			t.Errorf("%s: blesses %v, drops totems %v", b.ID, blesses(b), dropsTotems(b))
		}
		if s.PullSunder != (b.Class == proto.Class_ClassWarrior) || s.FlametongueTotem != (b.Spec == "Enhancement") {
			t.Errorf("%s: pull Sunder %v, Flametongue %v", b.ID, s.PullSunder, s.FlametongueTotem)
		}
	}
}

func TestStaleness(t *testing.T) {
	p := loadTestPool(t)
	stamps := p.Stamps()
	if s := p.StalenessOf(stamps); s.Stale() {
		t.Errorf("the pool is stale against itself: %s", s)
	}
	stamps["mage__fire"] = "old"
	stamps["mage__gone"] = "x"
	delete(stamps, "warlock__affliction")
	s := p.StalenessOf(stamps)
	if !slices.Equal(s.Changed, []string{"mage__fire"}) || !slices.Equal(s.Gone, []string{"mage__gone"}) ||
		!slices.Equal(s.Added, []string{"warlock__affliction"}) {
		t.Errorf("staleness %s", s)
	}
}
