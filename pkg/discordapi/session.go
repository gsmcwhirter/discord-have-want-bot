package discordapi

import (
	"sync"

	"github.com/gsmcwhirter/eso-discord/pkg/discordapi/etfapi"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
	"github.com/gsmcwhirter/eso-discord/pkg/snowflake"
	"github.com/pkg/errors"
)

// Session TODOC
type Session struct {
	lock      sync.RWMutex
	sessionID string
	state     etfapi.State

	guildParsers map[snowflake.Snowflake]parser.Parser
}

// ID TODOC
func (s *Session) ID() string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s == nil {
		return ""
	}
	return s.sessionID
}

// UpsertGuildFromElement TODOC
func (s *Session) UpsertGuildFromElement(e etfapi.Element) (err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err = s.state.UpsertGuildFromElement(e)
	return
}

// UpsertGuildFromElementMap TODOC
func (s *Session) UpsertGuildFromElementMap(eMap map[string]etfapi.Element) (err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	err = s.state.UpsertGuildFromElementMap(eMap)
	return
}

// UpsertChannelFromElement TODOC
func (s *Session) UpsertChannelFromElement(e etfapi.Element) (err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.state.UpsertChannelFromElement(e)

	return
}

// UpdateFromReady TODOC
func (s *Session) UpdateFromReady(p *etfapi.Payload) (err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	e, ok := p.Data["session_id"]
	if !ok {
		err = errors.Wrap(etfapi.ErrMissingData, "missing session_id")
		return
	}

	s.sessionID, err = e.ToString()
	if err != nil {
		err = errors.Wrap(err, "could not inflate session_id")
		return
	}

	s.state.UpdateFromReady(p)

	return
}
