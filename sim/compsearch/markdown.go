package compsearch

import (
	"fmt"
	"strings"

	"github.com/wowsims/classic/sim/core/proto"
)

func presetName(id proto.RaidPreset) string {
	if p := PresetByID(id); p != nil {
		return p.Name
	}
	return id.String()
}

func shortEnum(s, prefix string) string { return strings.TrimPrefix(s, prefix) }

// Markdown is a search result as a report.
func Markdown(res *proto.CompSearchResult) string {
	var b strings.Builder
	req := res.Request
	fmt.Fprintf(&b, "# Comp search: %d players, %d tanks, %d healers\n\n", req.Size, req.Tanks, req.Healers)
	e := req.Encounter
	if e.UseHealth {
		fmt.Fprintf(&b, "%s boss with %.0f health (the fight lasts until it dies)", e.Name, e.BossHealth)
	} else {
		fmt.Fprintf(&b, "%s boss, %.0f ± %.0f s", e.Name, e.Duration, e.DurationVariation)
	}
	if e.MobType != proto.MobType_MobTypeUnknown {
		fmt.Fprintf(&b, ", %s", shortEnum(e.MobType.String(), "MobType"))
	}
	if e.NaturalFlaskZone {
		b.WriteString(", in the natural flask zone")
	} else {
		b.WriteString(", outside the natural flask zone")
	}
	fmt.Fprintf(&b, ". %d single-player sims at %d iterations (contenders at %d), %d from the cache, %.0f s. Sim version %s.\n\n",
		res.SimsRun, req.Iterations, req.ContenderIterations, res.CacheHits, res.Seconds, res.Version)

	for _, sc := range res.Scenarios {
		camps := "camps off"
		if sc.CampBuffs {
			camps = "camps on"
		}
		fmt.Fprintf(&b, "## %s, %s\n\n", sc.Faction, camps)
		if e.UseHealth && sc.FightSeconds > 0 {
			fmt.Fprintf(&b, "The boss dies in %.1f s.\n\n", sc.FightSeconds)
		}
		if len(sc.Comps) > 0 {
			best := sc.Comps[0]
			if best.Confirmed {
				fmt.Fprintf(&b, "**Best: %.1f raid DPS in the raid sims** (±%.1f at 95%%; score %.1f), %s.\n\n", best.RaidSimDps, 1.96*best.RaidSimSe, best.Dps, comboText(best.Combo))
			} else {
				fmt.Fprintf(&b, "**Best: score %.1f** (±%.1f at 95%%), %s.\n\n", best.Dps, 1.96*best.Se, comboText(best.Combo))
			}
			b.WriteString("A score is the sum of the members' single-player sims in their seats; a raid sim runs the whole raid.\n\n")
			if sc.TieSet > 1 {
				fmt.Fprintf(&b, "The first %d comps below are a tie (within 2 standard errors or 0.1%%).\n\n", sc.TieSet)
			}
			for i, c := range sc.Comps {
				if c.Confirmed {
					fmt.Fprintf(&b, "%d. **%.1f** in the raid sims (%s; score %.1f, %s): %s\n", i+1, c.RaidSimDps, raidGapText(c), c.Dps, gapText(c), compText(c))
				} else {
					fmt.Fprintf(&b, "%d. score **%.1f** (%s): %s\n", i+1, c.Dps, gapText(c), compText(c))
				}
			}
			b.WriteString("\n")
			b.WriteString(membersTable(best))
		}
		if len(sc.Roles) > 0 {
			b.WriteString("\n**Roles** at the winning set (re-filled without each, or with it when it's missing):\n\n| Role | Gain | Provider | Provider cost | Net |\n|---|---|---|---|---|\n")
			for _, r := range sc.Roles {
				if !r.Measured {
					fmt.Fprintf(&b, "| %s | | %s | | not measured |\n", r.Role, r.Provider)
				} else if r.InSet {
					fmt.Fprintf(&b, "| %s | %.1f | %s | %.1f | %.1f |\n", r.Role, r.Gain, r.Provider, r.ProviderCost, r.Net)
				} else {
					fmt.Fprintf(&b, "| %s (missing) | | value of adding | | %.1f |\n", r.Role, r.Net)
				}
			}
		}
		if len(sc.Fillers) > 0 {
			b.WriteString("\n**Fillers** under the winning set, without a package:\n\n| Build | DPS | ±SE | Assumption range | |\n|---|---|---|---|---|\n")
			for _, f := range sc.Fillers {
				x := ""
				if f.Interchangeable {
					x = "interchangeable with the best"
				}
				fmt.Fprintf(&b, "| %s | %.1f | %.1f | %.0f to %.0f | %s |\n", f.BuildId, f.Dps, f.Se, f.AssumptionLow, f.AssumptionHigh, x)
			}
		}
		if len(sc.MemberValues) > 0 && len(sc.Comps) > 0 {
			b.WriteString("\n**Member values** (one row per distinct member):\n\n| Member | Own DPS | Contribution | Replacement margin | Best replacement | vs the report |\n|---|---|---|---|---|---|\n")
			for _, mv := range sc.MemberValues {
				name := mv.BuildId
				if mv.Count > 1 {
					name = fmt.Sprintf("%s (×%d)", name, mv.Count)
				}
				fmt.Fprintf(&b, "| %s | %.1f | %.1f | %.1f | %s | %+.1f (%s) |\n", name, mv.OwnDps, mv.Contribution,
					mv.ReplacementMargin, mv.BestReplacement, mv.ReportDifference, strings.Join(mv.ReportDifferenceParts, "; "))
			}
		}
		if len(sc.Adding) > 0 {
			b.WriteString("\n**Value of adding** a build that isn't in the comp (swapped for the member it beats most):\n\n")
			for i, a := range sc.Adding {
				if i >= 8 {
					break
				}
				fmt.Fprintf(&b, "- %s: %+.1f\n", a.BuildId, a.Value)
			}
		}
		if len(sc.Combos) > 0 {
			b.WriteString("\n**Tank and healer combos** (best first, by their best comp's screening score")
			if req.MaxCombos > 0 {
				fmt.Fprintf(&b, "; at most %d combos were ranked, so the preset values are over those", req.MaxCombos)
			}
			b.WriteString("):\n\n| Combo | Best comp's score | Behind |\n|---|---|---|\n")
			for i, c := range sc.Combos {
				if i >= 10 {
					fmt.Fprintf(&b, "| … %d more | | |\n", len(sc.Combos)-i)
					break
				}
				fmt.Fprintf(&b, "| %s | %.1f | %.1f |\n", comboText(c.Presets), c.Dps, c.Gap)
			}
			b.WriteString("\n**Each preset's DPS value**: at the winner's combo (its screening score less the best with that slot changed), and the best combo with it less the best without it:\n\n| Preset | At the winner's combo | With minus without |\n|---|---|---|\n")
			for _, p := range sc.PresetValues {
				fmt.Fprintf(&b, "| %s | %s | %s |\n", presetName(p.Preset), orNA(p.HasAtBest, p.AtBest), orNA(p.HasBestWithMinusWithout, p.BestWithMinusWithout))
			}
		}
		if len(sc.Buffs) > 0 {
			b.WriteString("\n**Buffs and debuffs**:\n\n")
			for _, r := range sc.Buffs {
				how := "given"
				if !r.Given {
					how = "cast"
				}
				fmt.Fprintf(&b, "- %s (%s, %s, %d members): %s\n", r.Effect, r.Scope, how, r.Receivers, r.Provider)
			}
		}
		if len(sc.Misses) > 0 {
			b.WriteString("\n**What the raid goes without**: " + strings.Join(sc.Misses, "; ") + ".\n")
		}
		if len(sc.Sensitivity) > 0 {
			var holds []string
			changes := 0
			for _, s := range sc.Sensitivity {
				if s.WinnerHolds {
					holds = append(holds, s.Scenario)
				} else {
					changes++
				}
			}
			fmt.Fprintf(&b, "\n**Sensitivity** to the report's open assumptions (each build shifted by its assumption runs, no sims): the winner holds under %d of %d.", len(holds), len(sc.Sensitivity))
			if changes > 0 {
				b.WriteString(" It changes under:\n\n")
				for _, s := range sc.Sensitivity {
					if !s.WinnerHolds && len(s.Winners) > 0 {
						if s.Winners[0].Gap != 0 {
							fmt.Fprintf(&b, "- %s: %s, ahead by %.1f\n", s.Scenario, s.Winners[0].BuffSet, s.Winners[0].Gap)
						} else {
							fmt.Fprintf(&b, "- %s (%s): %s wins, at %.1f\n", s.Scenario, s.Note, s.Winners[0].BuffSet, s.Winners[0].Dps)
						}
					}
				}
			} else {
				b.WriteString("\n")
			}
		}
		if mc := sc.ModelCheck; mc != nil {
			fmt.Fprintf(&b, "\nThe model scored %d buff sets; %d were simmed exactly, where it missed by %.1f at most (%.1f on average).\n",
				mc.SetsScored, mc.SetsSimmed, mc.MaxMiss, mc.MeanMiss)
		}
		for _, n := range sc.Notes {
			fmt.Fprintf(&b, "\n%s.\n", n)
		}
		if len(sc.Timings) > 0 {
			fmt.Fprintf(&b, "\nSims and seconds by phase: %s.\n", strings.Join(sc.Timings, ", "))
		}
		b.WriteString("\n")
	}
	b.WriteString("## Assumptions\n\n")
	for _, a := range res.Assumptions {
		fmt.Fprintf(&b, "- %s\n", a)
	}
	b.WriteString("\n**Left out** (no build can do it yet): " + strings.Join(res.LeftOut, "; ") + ".\n\n")
	b.WriteString("**Not modelled**: " + strings.Join(res.NotModelled, ", ") + ".\n")
	return b.String()
}

func gapText(c *proto.CompResult) string {
	if c.Gap == 0 {
		return "the best score"
	}
	return fmt.Sprintf("%.1f behind the best score, ±%.1f at 95%%", c.Gap, 1.96*c.GapSe)
}

func raidGapText(c *proto.CompResult) string {
	if c.RaidSimGap == 0 && c.RaidSimGapSe == 0 {
		return "the best"
	}
	return fmt.Sprintf("%.1f behind the best, ±%.1f at 95%%", c.RaidSimGap, 1.96*c.RaidSimGapSe)
}

func comboText(ids []proto.RaidPreset) string {
	var parts []string
	for _, id := range ids {
		parts = append(parts, presetName(id))
	}
	return strings.Join(parts, ", ")
}

func dpsMembers(c *proto.CompResult) []*proto.CompResultMember {
	var out []*proto.CompResultMember
	for _, m := range c.Members {
		if m.Role == proto.RaidRole_RaidRoleDps {
			out = append(out, m)
		}
	}
	return out
}

func compText(c *proto.CompResult) string {
	count := map[string]int{}
	var order []string
	for _, m := range dpsMembers(c) {
		id := m.BuildId
		if len(m.Duties) > 0 {
			id += " (" + shortEnum(m.Duties[0].String(), "Duty") + ")"
		}
		if count[id] == 0 {
			order = append(order, id)
		}
		count[id]++
	}
	var parts []string
	for _, id := range order {
		parts = append(parts, fmt.Sprintf("%d× %s", count[id], id))
	}
	return strings.Join(parts, ", ") + "; " + comboText(c.Combo)
}

func membersTable(c *proto.CompResult) string {
	var b strings.Builder
	b.WriteString("| Party | Member | Race | DPS | Duties and choices | Gets |\n|---|---|---|---|---|---|\n")
	for _, m := range c.Members {
		var jobs, gets []string
		for _, d := range m.Duties {
			job := shortEnum(d.String(), "Duty")
			if m.DutyCost != 0 {
				job += fmt.Sprintf(" (costs %.1f)", m.DutyCost)
			}
			jobs = append(jobs, job)
		}
		if m.Judgement != proto.PaladinJudgement_PaladinJudgementNone {
			jobs = append(jobs, "judges "+shortEnum(m.Judgement.String(), "PaladinJudgement"))
		}
		for _, t := range m.Totems {
			jobs = append(jobs, shortEnum(t.String(), "RaidTotem"))
		}
		for _, bl := range m.Blessings {
			gets = append(gets, shortEnum(bl.String(), "Blessing"))
		}
		if m.Package != proto.RaidPackage_RaidPackageNone {
			gets = append(gets, shortEnum(m.Package.String(), "RaidPackage"))
		}
		if m.PowerInfusion {
			gets = append(gets, "Power Infusion")
		}
		if m.Innervate {
			gets = append(gets, "Innervate")
		}
		dps := ""
		if m.Role == proto.RaidRole_RaidRoleDps {
			dps = fmt.Sprintf("%.1f", m.Dps)
		}
		fmt.Fprintf(&b, "| %d | %s | %s | %s | %s | %s |\n", m.Party, m.BuildId, m.Race, dps, strings.Join(jobs, ", "), strings.Join(gets, ", "))
	}
	return b.String()
}

func orNA(ok bool, v float64) string {
	if !ok {
		return "n/a"
	}
	return fmt.Sprintf("%.1f", v)
}
