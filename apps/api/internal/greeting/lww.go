package greeting

// Wins reports whether incoming should replace current using last-write-wins:
// later updatedAt wins; ties break on updatedBy then version so two devices
// with the same clock still converge.
func Wins(incoming, current Record) bool {
	if incoming.UpdatedAt.After(current.UpdatedAt) {
		return true
	}
	if incoming.UpdatedAt.Before(current.UpdatedAt) {
		return false
	}
	if incoming.UpdatedBy != current.UpdatedBy {
		return incoming.UpdatedBy > current.UpdatedBy
	}
	return incoming.Version > current.Version
}
