package storage

//go:generate protoc --go_out=. --proto_path=. ./userapi.proto

// UserAPI is the api for managing users transactions
type UserAPI interface {
	NewTransaction(writable bool) (UserAPITx, error)
}

// UserAPITx is the api for managing users within a transaction
type UserAPITx interface {
	Commit() error
	Rollback() error

	GetUser(name string) (User, error)
	GetUsers() []User

	AddUser(name string) (User, error)
	SaveUser(user User) error
}

// User is the api for managing a particular user
type User interface {
	GetName() string
	GetCharacter(name string) (Character, error)
	GetCharacters() []Character

	SetName(name string)
	AddCharacter(name string) Character
	DeleteCharacter(name string)

	Serialize() ([]byte, error)
}

// Character is the api for managing a user's particular character
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

// Skill is the api for managing a character's skill entry
type Skill interface {
	Name() string
	Points() uint64
}

// Item is theapi for managing a character's item entry
type Item interface {
	Name() string
	Count() uint64
}
