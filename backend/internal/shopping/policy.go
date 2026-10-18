package shopping

// CanAccessList returns whether a participant is authorized to view or mutate the list and history.
func CanAccessList(isParticipant bool) bool {
	return isParticipant
}
