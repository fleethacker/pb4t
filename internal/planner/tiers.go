package planner

// SuggestedTier provides illustrative planning priorities for legacy/demo events.
// Tier 1 is critical operations; 2 is major readiness training; 3 is routine
// checks; and 4 is administrative or support activity. These are editable guesses.
func SuggestedTier(eventType string) int {
	switch eventType {
	case "flight", "flight-operations", "diving", "aloft", "ammo-movement", "emcon":
		return 1
	case "damage-control", "dc-fire", "engineering-casualty", "loss-of-steering", "mass-casualty", "combat-exercise", "cstt-drill", "weapons-drill":
		return 2
	case "readiness", "chow":
		return 4
	default:
		return 3
	}
}
