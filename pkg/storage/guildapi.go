package storage

//go:generate protoc --go_out=. --proto_path=. ./guildapi.proto

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// ErrBadSetting is the error returned if an unknown setting is accessed
var ErrBadSetting = errors.New("bad setting")

// GuildSettings is the configuration settings set for a guild
type GuildSettings struct {
	ControlSequence string
}

// PrettyString returns a multi-line string representation of the guild settings
func (s *GuildSettings) PrettyString() string {
	return fmt.Sprintf(`
%[1]s
GuildSettings{
	ControlSequence: '%[2]s',
}
%[1]s
	`, "```", s.ControlSequence)
}

// GetSettingString returns the value of the requested setting
func (s *GuildSettings) GetSettingString(name string) (string, error) {
	switch strings.ToLower(name) {
	case "controlsequence":
		return s.ControlSequence, nil
	default:
		return "", ErrBadSetting
	}
}

// SetSettingString sets the value of the requested setting
func (s *GuildSettings) SetSettingString(name, val string) error {
	switch strings.ToLower(name) {
	case "controlsequence":
		s.ControlSequence = val
		return nil
	default:
		return ErrBadSetting
	}
}

// GuildAPI is the api for managing guilds transactions
type GuildAPI interface {
	NewTransaction(writable bool) (GuildAPITx, error)
}

// GuildAPITx is the api for managing guilds within a transaction
type GuildAPITx interface {
	Commit() error
	Rollback() error

	GetGuild(name string) (Guild, error)
	AddGuild(name string) (Guild, error)
	SaveGuild(guild Guild) error
}

// Guild is the api for managing a particular guild
type Guild interface {
	GetName() string
	GetSettings() GuildSettings

	SetName(name string)
	SetSettings(s GuildSettings)

	Serialize() ([]byte, error)
}
