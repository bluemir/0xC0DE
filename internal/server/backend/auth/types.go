package auth

const (
	KindUser           = "user"
	KindGroup          = "group"
	KindServiceAccount = "service-account"
	KindGuest          = "guest"
)

type Verb string
type Resource = KeyValues

type KeyValues map[string]string

func (kvs KeyValues) Get(key string) string {
	return kvs[key]
}

func (kvs KeyValues) IsSubsetOf(resource Resource) bool {
	for k, v := range kvs {
		if v == "*" {
			continue
		}
		if resource[k] != v {
			return false
		}
	}
	return true
}

type Labels map[string]string

type Context struct {
	User     *User    `expr:"user"`
	Subject  Subject  `expr:"subject"`
	Verb     Verb     `expr:"verb"`
	Resource Resource `expr:"resource"`
}
