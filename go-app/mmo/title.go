package mmo

func (p *Player) grantTitle(id string) {
	if titleByID[id].ID == "" {
		return
	}
	log := p.ensureLog()
	for _, t := range log.Titles {
		if t == id {
			return
		}
	}
	log.Titles = append(log.Titles, id)
	if log.ActiveTitle == "" {
		log.ActiveTitle = id
		p.Title = titleByID[id].Name
	}
	p.markDirty()
}

func (w *WorldState) ApplyTitle(id string, env Envelope) [][]byte {
	p := w.players[id]
	if p == nil {
		return nil
	}
	var in struct{ ID string }
	_ = unmarshal(env.Data, &in)
	switch env.Type {
	case TypeSetTitle:
		owned := false
		for _, t := range p.ensureLog().Titles {
			if t == in.ID {
				owned = true
			}
		}
		if !owned || titleByID[in.ID].ID == "" {
			return rejectFor(p.ID, TypeSetTitle, "owned")
		}
		p.ensureLog().ActiveTitle = in.ID
		p.Title = titleByID[in.ID].Name
		w.persist(p)
		return [][]byte{marshal(TypeFriendUpdated, w.socialState(p))}
	case TypeSetCosmetic:
		owned := false
		for _, c := range p.ensureLog().Cosmetics {
			if c == in.ID {
				owned = true
			}
		}
		if !owned {
			return rejectFor(p.ID, TypeSetCosmetic, "owned")
		}
		p.ensureLog().ActiveCosmetic = in.ID
		w.persist(p)
		return [][]byte{marshal(TypeFriendUpdated, w.socialState(p))}
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) refreshTitles(p *Player) {
	log := p.ensureLog()
	if log.DiscoveredZones["forest"] {
		p.grantTitle("forest-explorer")
	}
	if len(log.Guardians) > 0 {
		p.grantTitle("guardian-hunter")
	}
	if log.Flags["event_village-defense_done"] {
		p.grantTitle("dawn-defender")
	}
	if log.Flags["celestial_gate_unlocked"] || log.DiscoveredZones["celestial"] {
		p.grantTitle("world-adventurer")
	}
	if log.ActiveTitle != "" {
		p.Title = titleByID[log.ActiveTitle].Name
	}
}
