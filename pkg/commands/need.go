package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/discord-bot-lib/util"
	"github.com/gsmcwhirter/go-util/parser"

	"github.com/gsmcwhirter/discord-have-want-bot/pkg/storage"
)

type needItemHandler struct {
	user     storage.User
	charName string
}

func (h *needItemHandler) HandleLine(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleEmbedResponse{
		To: user,
	}

	args, ctRunes := parser.MaybeCount(args)

	itemName := strings.TrimSpace(args)

	if len(itemName) == 0 {
		return r, ErrItemNameRequired
	}

	ctStr := strings.TrimSpace(ctRunes)
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

	char.IncrNeededItem(itemName, uint64(ct))

	r.Description = fmt.Sprintf("marked %s as needing +%d of %s", h.charName, ct, itemName)
	return r, nil
}

type needPointsHandler struct {
	user     storage.User
	charName string
}

func (h *needPointsHandler) HandleLine(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleEmbedResponse{
		To: user,
	}

	args, ctRunes := parser.MaybeCount(args)

	skillName := strings.TrimSpace(args)

	if len(skillName) == 0 {
		return r, ErrSkillNameRequired
	}

	ctStr := strings.TrimSpace(ctRunes)
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

	char.IncrNeededSkill(skillName, uint64(ct))

	r.Description = fmt.Sprintf("marked %s as needing +%d points in %s", h.charName, ct, skillName)
	return r, nil
}

type needCommands struct {
	preCommand string
	deps       dependencies
}

func (c *needCommands) helpCharsPoints(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: user,
	}

	r.Description = fmt.Sprintf("Usage: %s [%s] [skill name] [count?]\n\n", c.preCommand+" pts", "charname")

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, nil
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
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

func (c *needCommands) helpCharsItems(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: user,
	}

	r.Description = fmt.Sprintf("Usage: %s [%s] [item name] [count?]\n\n", c.preCommand+" pts", "charname")

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, nil
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
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

func (c *needCommands) points(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleEmbedResponse{
		To: user,
	}

	t, err := c.deps.UserAPI().NewTransaction(true)
	if err != nil {
		return r, err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		bUser, err = t.AddUser(user)
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
	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  c.preCommand + " pts",
		Placeholder: "charname",
	})
	ch.SetHandler("", cmdhandler.NewLineHandler(c.helpCharsPoints))
	for _, char := range characters {
		ch.SetHandler(char.GetName(), &needPointsHandler{charName: char.GetName(), user: bUser})
	}

	r2, err := ch.HandleLine(user, guild, args)

	if err != nil {
		return r2, err
	}

	err = t.SaveUser(bUser)
	if err != nil {
		return r2, errors.Wrap(err, "could not save points need")
	}

	err = t.Commit()
	if err != nil {
		return r2, errors.Wrap(err, "could not save points need")
	}

	return r2, nil
}

func (c *needCommands) item(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleEmbedResponse{
		To: user,
	}

	t, err := c.deps.UserAPI().NewTransaction(true)
	if err != nil {
		return r, err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		bUser, err = t.AddUser(user)
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
	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  c.preCommand + " item",
		Placeholder: "charname",
	})
	ch.SetHandler("", cmdhandler.NewLineHandler(c.helpCharsItems))
	for _, char := range characters {
		ch.SetHandler(char.GetName(), &needItemHandler{charName: char.GetName(), user: bUser})
	}
	r2, err := ch.HandleLine(user, guild, args)

	if err != nil {
		return r2, err
	}

	err = t.SaveUser(bUser)
	if err != nil {
		return r2, errors.Wrap(err, "could not save item need")
	}

	err = t.Commit()
	if err != nil {
		return r2, errors.Wrap(err, "could not save item need")
	}

	return r2, nil
}

// NeedCommandHandler TODOC
func NeedCommandHandler(deps dependencies, preCommand string) *cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: " ",
	})
	nc := needCommands{
		preCommand: preCommand,
		deps:       deps,
	}
	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  preCommand,
		Placeholder: "type",
	})
	ch.SetHandler("pts", cmdhandler.NewLineHandler(nc.points))
	ch.SetHandler("item", cmdhandler.NewLineHandler(nc.item))

	return ch
}
