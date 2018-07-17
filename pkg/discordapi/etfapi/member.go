package etfapi

import (
	"github.com/pkg/errors"

	"github.com/gsmcwhirter/eso-discord/pkg/snowflake"
)

// GuildMember TODOC
type GuildMember struct {
	id    snowflake.Snowflake
	roles []string
	user  User
}

// GuildMemberFromElement TODOC
func GuildMemberFromElement(e Element) (m GuildMember, err error) {

	eMap, err := e.ToMap()
	if err != nil {
		err = errors.Wrap(err, "could not inflate guild member to element map")
		return
	}

	m.user, err = UserFromElement(eMap["user"])
	if err != nil {
		err = errors.Wrap(err, "could not inflate guild member user")
		return
	}
	m.id = m.user.id

	// TODO roles

	return
}
