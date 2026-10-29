package compsearch

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wowsims/classic/sim/core/proto"
)

// The pool needs the item database: go test -tags with_db.

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}

var testPool *Pool

func loadTestPool(t testing.TB) *Pool {
	t.Helper()
	if testPool == nil {
		p, err := LoadPool(repoRoot())
		if err != nil {
			t.Fatal(err)
		}
		testPool = p
	}
	return testPool
}

func TestPoolGifts(t *testing.T) {
	p := loadTestPool(t)
	if len(p.Bases) != 37 {
		t.Errorf("%d base builds, want 37", len(p.Bases))
	}
	want := map[string]Gift{
		"druid__balance":              GiftGotW | GiftMoonkin,
		"druid__feral-cat":            GiftGotW | GiftLotP,
		"druid__feral-cat-duty-ff":    GiftGotW | GiftLotP | GiftMinorArmor,
		"hunter__marksmanship":        GiftTrueshot,
		"hunter__beast-mastery":       0,
		"mage__fire":                  GiftAB,
		"priest__shadow":              GiftPoS | GiftFort,
		"paladin__level_60":           GiftBlessing,
		"shaman__enhancement-level60": GiftSoE | GiftWF,
		"shaman__elemental-level60":   0,
		"warrior__fury":               0,
		"warrior__fury-duty-battle-shout": GiftBattleShout,
		"warlock__demonology-duty-curse":  GiftCoE,
		"rogue__subtlety":                 0,
	}
	for id, g := range want {
		b := p.Builds[id]
		if b == nil {
			t.Errorf("no build %s", id)
			continue
		}
		if b.Gives != g {
			t.Errorf("%s gives %v, want %v", id, b.Gives, g)
		}
	}
	for _, b := range p.Builds {
		for _, f := range []proto.Faction{proto.Faction_Horde, proto.Faction_Alliance} {
			if _, err := p.Player(b, f); err != nil {
				t.Errorf("%s %s: %v", b.ID, f, err)
			}
		}
	}
	if p.Builds["hunter__beast-mastery"].Race[proto.Faction_Alliance] != proto.Race_RaceNightElf {
		t.Errorf("Beast Mastery's Alliance race: %v", p.Builds["hunter__beast-mastery"].Race[proto.Faction_Alliance])
	}
	if p.Builds["shaman__elemental-level60-mana-tide"].Pkg != proto.RaidPackage_RaidPackageManaSpringTide {
		t.Errorf("the + Mana Tide build's package")
	}
}
