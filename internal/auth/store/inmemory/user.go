package static

import "github.com/bluemir/0xC0DE/internal/auth"

func (store *Store) CreateUser(user *auth.User) error {
	store.users[user.Name] = *user
	return nil
}
func (store *Store) GetUser(username string) (*auth.User, error) {
	user := store.users[username]
	return &user, nil
}
func (store *Store) ListUser() ([]auth.User, error) {
	users := []auth.User{}
	for _, user := range store.users {
		users = append(users, user)
	}
	return users, nil
}
func (store *Store) UpdateUser(user *auth.User) error {
	store.users[user.Name] = *user
	return nil
}
func (store *Store) DeleteUser(username string) error {
	delete(store.users, username)
	return nil
}
