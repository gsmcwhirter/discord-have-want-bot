package storage

//go:generate protoc --go_out=. --proto_path=. ./guildapi.proto

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// ErrBadSetting TODOC
var ErrBadSetting = errors.New("bad setting")

// GuildSettings TODOC
type GuildSettings struct {
	ControlSequence string
}

// PrettyString TODOC
func (s *GuildSettings) PrettyString() string {
	return fmt.Sprintf(`
%[1]s
GuildSettings{
	ControlSequence: '%[2]s',
}
%[1]s
	`, "```", s.ControlSequence)
}

// GetSettingString TODOC
func (s *GuildSettings) GetSettingString(name string) (string, error) {
	switch strings.ToLower(name) {
	case "controlsequence":
		return s.ControlSequence, nil
	default:
		return "", ErrBadSetting
	}
}

// SetSettingString TODOC
func (s *GuildSettings) SetSettingString(name, val string) error {
	switch strings.ToLower(name) {
	case "controlsequence":
		s.ControlSequence = val
		return nil
	default:
		return ErrBadSetting
	}
}

// GuildAPI TODOC
type GuildAPI interface {
	NewTransaction(writable bool) (GuildAPITx, error)
}

// GuildAPITx TODOC
type GuildAPITx interface {
	Commit() error
	Rollback() error

	GetGuild(name string) (Guild, error)
	AddGuild(name string) (Guild, error)
	SaveGuild(guild Guild) error
}

// Guild TODOC
type Guild interface {
	GetName() string
	GetSettings() GuildSettings

	SetName(name string)
	SetSettings(s GuildSettings)

	Serialize() ([]byte, error)
}
