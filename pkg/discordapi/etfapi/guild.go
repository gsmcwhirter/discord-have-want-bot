package etfapi

import (
	"github.com/pkg/errors"

	"github.com/gsmcwhirter/eso-discord/pkg/snowflake"
)

// Guild TODOC
type Guild struct {
	id            snowflake.Snowflake
	ownerID       snowflake.Snowflake
	applicationID snowflake.Snowflake
	name          string
	available     bool
	members       []GuildMember
	channels      []Channel
}

// ID TODOC
func (g *Guild) ID() snowflake.Snowflake {
	return g.id
}

// UpdateFromElementMap TODOC
func (g *Guild) UpdateFromElementMap(eMap map[string]Element) (err error) {
	var ok bool
	var e2 Element
	var m GuildMember
	var c Channel

	e2, ok = eMap["owner_id"]
	if ok && !e2.IsNil() {
		g.ownerID, err = SnowflakeFromElement(e2)
		if err != nil {
			err = errors.Wrap(err, "could not get owner_id snowflake.Snowflake")
			return
		}
	}

	e2, ok = eMap["application_id"]
	if ok && !e2.IsNil() {
		g.applicationID, err = SnowflakeFromElement(e2)
		if err != nil {
			err = errors.Wrap(err, "could not get application_id snowflake.Snowflake")
			return
		}
	}

	e2, ok = eMap["name"]
	if ok {
		g.name, err = e2.ToString()
		if err != nil {
			err = errors.Wrap(err, "could not get name")
			return
		}
	}

	e2, ok = eMap["unavailable"]
	if ok {
		var uavStr string
		uavStr, err = e2.ToString()
		if err != nil {
			err = errors.Wrap(err, "could not get unavailable status")
			return
		}

		switch uavStr {
		case "true":
			g.available = false
		case "false":
			g.available = true
		default:
			err = errors.Wrap(ErrBadData, "did not get true or false availability")
			return
		}
	}

	e2, ok = eMap["members"]
	if ok {
		g.members = make([]GuildMember, 0, len(e2.Vals))
		for _, e3 := range e2.Vals {
			m, err = GuildMemberFromElement(e3)
			if err != nil {
				err = errors.Wrap(err, "could not inflate guild member")
				return
			}
			g.members = append(g.members, m)
		}
	}

	e2, ok = eMap["channels"]
	if ok {
		g.channels = make([]Channel, 0, len(e2.Vals))
		for _, e3 := range e2.Vals {
			c, err = ChannelFromElement(e3)
			if err != nil {
				err = errors.Wrap(err, "could not inflate guild channel")
				return
			}
			g.channels = append(g.channels, c)
		}
	}
	return
}

// GuildFromElement TODOC
func GuildFromElement(e Element) (g Guild, err error) {
	var eMap map[string]Element
	eMap, g.id, err = MapAndIDFromElement(e)
	if err != nil {
		return
	}

	err = g.UpdateFromElementMap(eMap)
	err = errors.Wrap(err, "could not create a guild")

	return
}
