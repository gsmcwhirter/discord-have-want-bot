package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/go-util/deferutil"
	"github.com/gsmcwhirter/go-util/parser"

	"github.com/gsmcwhirter/discord-have-want-bot/pkg/storage"
)

type gotItemHandler struct {
	user     storage.User
	charName string
}

func (h *gotItemHandler) HandleMessage(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleEmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	args, ctStr := parser.MaybeCount(msg.Contents())

	itemName := strings.TrimSpace(args)

	if len(itemName) == 0 {
		return r, ErrItemNameRequired
	}

	ctStr = strings.TrimSpace(ctStr)
	if ctStr == "" {
		ctStr = "1"
	}

	ct, err := strconv.Atoi(ctStr)
	if err != nil {
		return r, errors.Wrap(err, "could not interpret count to adjust item needs")
	}

	if ct < 0 {
		return r, ErrPositiveValueRequired
	}

	char, err := h.user.GetCharacter(h.charName)
	if err != nil {
		return r, errors.Wrap(err, "could not find character to adjust item needs")
	}

	char.DecrNeededItem(itemName, uint64(ct))

	r.Description = fmt.Sprintf("marked %s as needing -%d of %s", h.charName, ct, itemName)
	return r, nil
}

type gotPointsHandler struct {
	user     storage.User
	charName string
}

func (h *gotPointsHandler) HandleMessage(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleEmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	args, ctStr := parser.MaybeCount(msg.Contents())

	skillName := strings.TrimSpace(args)

	if len(skillName) == 0 {
		return r, ErrSkillNameRequired
	}

	ctStr = strings.TrimSpace(ctStr)
	if ctStr == "" {
		ctStr = "1"
	}

	ct, err := strconv.Atoi(ctStr)
	if err != nil {
		return r, errors.Wrap(err, "could not interpret count to adjust skill needs")
	}

	if ct < 0 {
		return r, ErrPositiveValueRequired
	}

	char, err := h.user.GetCharacter(h.charName)
	if err != nil {
		return r, errors.Wrap(err, "could not find character to adjust skill needs")
	}

	char.DecrNeededSkill(skillName, uint64(ct))

	r.Description = fmt.Sprintf("marked %s as needing -%d points in %s", h.charName, ct, skillName)
	return r, nil
}

type gotCommands struct {
	preCommand string
	deps       dependencies
}

func (c *gotCommands) helpCharsPoints(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	r.Description = fmt.Sprintf("Usage: %s [%s] [skill name] [count?]\n\n", c.preCommand+" pts", "charname")

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, nil
	}
	defer deferutil.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(msg.UserID().ToString())
	if err != nil {
		return r, nil
	}

	characters := bUser.GetCharacters()
	charNames := make([]string, 0, len(characters))

	for _, char := range characters {
		charNames = append(charNames, char.GetName())
	}

	sort.Strings(charNames)
	r.Fields = []cmdhandler.EmbedField{
		{
			Name: "*Available Character Names*",
			Val:  fmt.Sprintf("```\n%s\n```\n", strings.Join(charNames, "\n")),
		},
	}

	return r, nil
}

func (c *gotCommands) helpCharsItems(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	r.Description = fmt.Sprintf("Usage: %s [%s] [item name] [count?]\n\n", c.preCommand+" item", "charname")

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, nil
	}
	defer deferutil.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(msg.UserID().ToString())
	if err != nil {
		return r, nil
	}

	characters := bUser.GetCharacters()
	charNames := make([]string, 0, len(characters))

	for _, char := range characters {
		charNames = append(charNames, char.GetName())
	}

	sort.Strings(charNames)
	r.Fields = []cmdhandler.EmbedField{
		{
			Name: "*Available Character Names*",
			Val:  fmt.Sprintf("```\n%s\n```\n", strings.Join(charNames, "\n")),
		},
	}

	return r, nil
}

func (c *gotCommands) points(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleEmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	t, err := c.deps.UserAPI().NewTransaction(true)
	if err != nil {
		return r, err
	}
	defer deferutil.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(msg.UserID().ToString())
	if err != nil {
		bUser, err = t.AddUser(msg.UserID().ToString())
		if err != nil {
			return r, errors.Wrap(err, "could not create user")
		}
	}

	characters := bUser.GetCharacters()
	charNames := make([]string, len(characters))
	for i, char := range characters {
		charNames[i] = char.GetName()
	}

	p := parser.NewParser(parser.Options{
		CmdIndicator: " ",
	})
	ch, err := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  c.preCommand + " pts",
		Placeholder: "charname",
	})
	if err != nil {
		return r, err
	}

	ch.SetHandler("", cmdhandler.NewMessageHandler(c.helpCharsPoints))
	for _, char := range characters {
		ch.SetHandler(char.GetName(), &gotPointsHandler{charName: char.GetName(), user: bUser})
	}

	r2, err := ch.HandleMessage(msg)

	if err != nil {
		return r2, err
	}

	err = t.SaveUser(bUser)
	if err != nil {
		return r2, errors.Wrap(err, "could not save points gotten")
	}

	err = t.Commit()
	if err != nil {
		return r2, errors.Wrap(err, "could not save points gotten")
	}

	return r2, nil
}

func (c *gotCommands) item(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleEmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	t, err := c.deps.UserAPI().NewTransaction(true)
	if err != nil {
		return r, err
	}
	defer deferutil.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(msg.UserID().ToString())
	if err != nil {
		bUser, err = t.AddUser(msg.UserID().ToString())
		if err != nil {
			return r, errors.Wrap(err, "could not create user")
		}
	}

	characters := bUser.GetCharacters()
	charNames := make([]string, len(characters))
	for i, char := range characters {
		charNames[i] = char.GetName()
	}

	p := parser.NewParser(parser.Options{
		CmdIndicator: " ",
	})
	ch, err := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  c.preCommand + " item",
		Placeholder: "charname",
	})
	if err != nil {
		return r, err
	}

	ch.SetHandler("", cmdhandler.NewMessageHandler(c.helpCharsItems))
	for _, char := range characters {
		ch.SetHandler(char.GetName(), &gotItemHandler{charName: char.GetName(), user: bUser})
	}
	r2, err := ch.HandleMessage(msg)

	if err != nil {
		return r2, err
	}

	err = t.SaveUser(bUser)
	if err != nil {
		return r2, errors.Wrap(err, "could not save item gotten")
	}

	err = t.Commit()
	if err != nil {
		return r2, errors.Wrap(err, "could not save item gotten")
	}

	return r2, nil
}

// GotCommandHandler creates a new command handler for !got commands
func GotCommandHandler(deps dependencies, preCommand string) (*cmdhandler.CommandHandler, error) {
	p := parser.NewParser(parser.Options{
		CmdIndicator: " ",
	})
	gc := gotCommands{
		preCommand: preCommand,
		deps:       deps,
	}
	ch, err := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:          preCommand,
		Placeholder:         "type",
		HelpOnEmptyCommands: true,
	})
	if err != nil {
		return nil, err
	}

	ch.SetHandler("pts", cmdhandler.NewMessageHandler(gc.points))
	ch.SetHandler("item", cmdhandler.NewMessageHandler(gc.item))

	return ch, nil
}
