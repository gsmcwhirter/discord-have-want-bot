package storage

import (
	"errors"

	"github.com/golang/protobuf/proto"
)

// ErrCharacterNotExist TODOC
var ErrCharacterNotExist = errors.New("character does not exist")

type boltUser struct {
	protoUser *ProtoUser
}

func (u *boltUser) GetName() string {
	return u.protoUser.Name
}

func (u *boltUser) GetCharacter(name string) (Character, error) {
	if u.protoUser.Characters == nil {
		return nil, ErrCharacterNotExist
	}

	protoChar, ok := u.protoUser.Characters[name]
	if !ok {
		return nil, ErrCharacterNotExist
	}

	return &boltCharacter{protoChar}, nil
}

func (u *boltUser) GetCharacters() []Character {
	if u.protoUser.Characters == nil {
		return []Character{}
	}

	chars := make([]Character, len(u.protoUser.Characters))
	i := 0
	for _, protoChar := range u.protoUser.Characters {
		chars[i] = &boltCharacter{protoChar}
		i++
	}
	return chars
}

func (u *boltUser) SetName(name string) {
	u.protoUser.Name = name
}

func (u *boltUser) AddCharacter(name string) Character {
	if u.protoUser.Characters == nil {
		u.protoUser.Characters = map[string]*ProtoCharacter{}
	}

	char, err := u.GetCharacter(name)
	if err != nil {
		protoChar := &ProtoCharacter{Name: name}
		u.protoUser.Characters[name] = protoChar
		char = &boltCharacter{protoChar}
	}
	return char
}

func (u *boltUser) DeleteCharacter(name string) {
	if u.protoUser.Characters == nil {
		return
	}

	_, ok := u.protoUser.Characters[name]
	if ok {
		delete(u.protoUser.Characters, name)
	}
}

func (u *boltUser) Serialize() (out []byte, err error) {
	out, err = proto.Marshal(u.protoUser)
	return
}
