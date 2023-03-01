package inmemory

import "github.com/bluemir/0xC0DE/internal/auth"

func (s *Store) CreateToken(token *auth.Token) error {
	s.tokens[TokenIndex{Username: token.Username, HashedKey: token.HashedKey}] = *token
	return nil
}
func (s *Store) GetToken(username, hashedKey string) (*auth.Token, error) {
	token := s.tokens[TokenIndex{Username: username, HashedKey: hashedKey}]
	return &token, nil
}
func (s *Store) ListToken(username string) ([]auth.Token, error) {
	tokens := []auth.Token{}
	for _, token := range s.tokens {
		tokens = append(tokens, token)
	}
	return tokens, nil
}
func (s *Store) DeleteToken(username, revokeKey string) error {
	for _, token := range s.tokens {
		if token.Username == username && token.RevokeKey == revokeKey {
			delete(s.tokens, TokenIndex{Username: username, HashedKey: token.HashedKey})
		}
	}
	return nil
}
