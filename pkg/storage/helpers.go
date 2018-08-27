package storage

import (
	"github.com/gsmcwhirter/discord-bot-lib/snowflake"
	"github.com/gsmcwhirter/go-util/deferutil"
	"github.com/pkg/errors"
)

// GetSettings gets the guild configuration settings
//
// NOTE: this cannot be called while another transaction is open
func GetSettings(gapi GuildAPI, gid snowflake.Snowflake) (s GuildSettings, err error) {
	t, err := gapi.NewTransaction(false)
	if err != nil {
		return
	}
	defer deferutil.CheckDefer(t.Rollback)

	bGuild, err := t.AddGuild(gid.ToString())
	if err != nil {
		err = errors.Wrap(err, "unable to find guild")
		return
	}

	s = bGuild.GetSettings()
	return
}
