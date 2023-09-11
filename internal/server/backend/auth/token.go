package auth

import "github.com/rs/xid"

func (m *Manager) IssueToken(username, unhashedKey string, opts ...TokenOpt) (*Token, error) {
	user, err := m.store.GetUser(username)
	if err != nil {
		return nil, err
	}

	token := &Token{
		Username:  username,
		HashedKey: Hash(unhashedKey, user.Salt, m.salt),
		RevokeKey: xid.New().String(),
	}

	if err := m.store.CreateToken(token); err != nil {
		return nil, err
	}
	return token, nil
}
func (m *Manager) GenerateToken(username string, opts ...TokenOpt) (string, error) {
	user, err := m.store.GetUser(username)
	if err != nil {
		return "", err
	}

	unhashedKey := Hash(xid.New().String(), m.salt)
	token := &Token{
		Username:  username,
		HashedKey: Hash(unhashedKey, user.Salt, m.salt),
		RevokeKey: xid.New().String(),
	}

	if err := m.store.CreateToken(token); err != nil {
		return unhashedKey, err
	}

	return unhashedKey, nil
}
func (m *Manager) GetToken(username, unhashedKey string) (*Token, error) {
	user, err := m.store.GetUser(username)
	if err != nil {
		return nil, err
	}
	hashedKey := Hash(unhashedKey, user.Salt, m.salt)

	return m.store.GetToken(username, hashedKey)
}
func (m *Manager) ListToken(username string) ([]Token, error) {
	return m.store.ListToken(username)
}
func (m *Manager) RevokeToken(username, revokeKey string) error {
	return m.store.DeleteToken(username, revokeKey)
}
