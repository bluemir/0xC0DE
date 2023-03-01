package auth

func NewManager(store AuthStore) (IManager, error) {
	return New(store, "salt")
}
