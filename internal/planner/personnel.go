package planner

type personnelSlot struct {
	e          Event
	rule       Rule
	candidates []Person
	blocked    []string
}

// competingPersonnel applies to two roles filled by the same person.
func competingPersonnel(d Data, a, b Event) bool {
	return a.ID == b.ID || Overlap(a.Start, a.End(), b.Start, b.End()) ||
		(a.Fatigue && b.Fatigue && tooClose(a.Start, a.End(), b.Start, b.End(), max(restFor(d, a), restFor(d, b))))
}

// personnelGroups keeps all roles for an event together and connects events
// only when a shared eligible person could be double-booked or miss rest.
// Independent groups can be staffed and reported independently.
func personnelGroups(d Data, slots []personnelSlot) [][]personnelSlot {
	var groups [][]personnelSlot
	visited := make([]bool, len(slots))
	for i, first := range slots {
		if visited[i] {
			continue
		}
		visited[i] = true
		group := []personnelSlot{first}
		for next := 0; next < len(group); next++ {
			current := group[next]
			for j, candidate := range slots {
				if visited[j] {
					continue
				}
				connected := current.e.ID == candidate.e.ID
				if !connected && competingPersonnel(d, current.e, candidate.e) {
					for _, a := range current.candidates {
						for _, b := range candidate.candidates {
							if a.ID == b.ID {
								connected = true
							}
						}
					}
				}
				if connected {
					visited[j] = true
					group = append(group, candidate)
				}
			}
		}
		groups = append(groups, group)
	}
	return groups
}
