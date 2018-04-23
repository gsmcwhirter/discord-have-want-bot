package storage

// UserAPI TODOC
type UserAPI interface {
	NewTransaction(writable bool) (UserAPITx, error)
}

// UserAPITx TODOC
type UserAPITx interface {
	Commit() error
	Rollback() error

	GetUser(name string) (User, error)
	GetUsers() []User

	AddUser(name string) (User, error)
	SaveUser(user User) error
}

// User TODOC
type User interface {
	GetName() string
	GetCharacter(name string) (Character, error)
	GetCharacters() []Character

	SetName(name string)
	AddCharacter(name string) Character
	DeleteCharacter(name string)

	Serialize() ([]byte, error)
}

// Character TODOC
type Character interface {
	GetName() string
	GetNeededSkill(name string) (Skill, error)
	GetNeededSkills() []Skill
	GetNeededItem(name string) (Item, error)
	GetNeededItems() []Item

	SetName(name string)
	IncrNeededSkill(name string, amt uint64)
	DecrNeededSkill(name string, amt uint64)
	IncrNeededItem(name string, amt uint64)
	DecrNeededItem(name string, amt uint64)
}

// Skill TODOC
type Skill interface {
	Name() string
	Points() uint64
}

// Item TODOC
type Item interface {
	Name() string
	Count() uint64
}
