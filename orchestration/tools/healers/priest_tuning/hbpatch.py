p = '.ps/hb/main.go'
s = open(p).read()
s = s.replace('''	logOne := flag.Bool("log", false, "print one iteration's log of the first build instead")''', '''	logOne := flag.Bool("log", false, "print one iteration's log of the first build instead")
	breakdown := flag.Bool("breakdown", false, "print the first healer's healing by spell")''')
s = s.replace('''		r := b.run(o.scale, false)
		row := append''', '''		r := b.run(o.scale, false)
		if *breakdown {
			b.breakdown(o.scale)
		}
		row := append''')
s += '''
func (b *bench) breakdown(scale float64) {
	req := b.request(scale, false)
	res := core.RunRaidSimConcurrent(req)
	n := float64(req.SimOptions.Iterations)
	for pi, party := range res.RaidMetrics.Parties {
		for si, m := range party.Players {
			if b.slots[pi][si] != roleHealer {
				continue
			}
			type row struct{ casts, healing, over, crits, hits float64 }
			rows := map[string]*row{}
			var names []string
			for _, a := range m.Actions {
				id := a.Id.String()
				if rows[id] == nil {
					rows[id] = &row{}
					names = append(names, id)
				}
				for _, t := range a.Targets {
					rows[id].casts += float64(t.Casts) / n
					rows[id].healing += t.Healing / n
					rows[id].over += t.Overhealing / n
					rows[id].crits += float64(t.Crits + t.CritTicks)
					rows[id].hits += float64(t.Hits + t.Ticks + t.Crits + t.CritTicks)
				}
			}
			for _, id := range names {
				r := rows[id]
				if r.casts == 0 && r.healing == 0 {
					continue
				}
				oh := 0.0
				if r.healing > 0 {
					oh = r.over / r.healing
				}
				fmt.Printf("    %-40s casts %6.1f  healing %8.0f  overheal %5.1f%%  crit %5.1f%%\\n", id, r.casts, r.healing, 100*oh, 100*r.crits/max(1, r.hits))
			}
			fmt.Printf("    mana spent %.0f, end %.0f\\n", m.ManaSpentAvg, m.ManaEnd.Avg)
			return
		}
	}
}
'''
open(p, 'w').write(s)
