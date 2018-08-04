package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/discord-bot-lib/util"
	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"

	"github.com/gsmcwhirter/eso-discord/pkg/storage"
)

type gotItemHandler struct {
	user     storage.User
	charName string
}

func (h *gotItemHandler) HandleLine(user, guild, args string) (string, error) {
	args, ctRunes := parser.MaybeCount(args)

	itemName := strings.TrimSpace(args)

	if len(itemName) == 0 {
		return "", ErrItemNameRequired
	}

	ctStr := strings.TrimSpace(ctRunes)
	if ctStr == "" {
		ctStr = "1"
	}

	ct, err := strconv.Atoi(ctStr)
	if err != nil {
		return "", errors.Wrap(err, "could not interpret count to adjust item needs")
	}

	if ct < 0 {
		return "", ErrPositiveValueRequired
	}

	char, err := h.user.GetCharacter(h.charName)
	if err != nil {
		return "", errors.Wrap(err, "could not find character to adjust item needs")
	}

	char.DecrNeededItem(itemName, uint64(ct))

	return fmt.Sprintf("marked %s as needing -%d of %s", h.charName, ct, itemName), nil
}

type gotPointsHandler struct {
	user     storage.User
	charName string
}

func (h *gotPointsHandler) HandleLine(user, guild, args string) (string, error) {
	args, ctRunes := parser.MaybeCount(args)

	skillName := strings.TrimSpace(args)

	if len(skillName) == 0 {
		return "", ErrSkillNameRequired
	}

	ctStr := strings.TrimSpace(ctRunes)
	if ctStr == "" {
		ctStr = "1"
	}

	ct, err := strconv.Atoi(ctStr)
	if err != nil {
		return "", errors.Wrap(err, "could not interpret count to adjust skill needs")
	}

	if ct < 0 {
		return "", ErrPositiveValueRequired
	}

	char, err := h.user.GetCharacter(h.charName)
	if err != nil {
		return "", errors.Wrap(err, "could not find character to adjust skill needs")
	}

	char.DecrNeededSkill(skillName, uint64(ct))

	return fmt.Sprintf("marked %s as needing -%d points in %s", h.charName, ct, skillName), nil
}

type gotCommands struct {
	preCommand string
	deps       dependencies
}

func (c *gotCommands) helpCharsPoints(user, guild, args string) (string, error) {
	helpStr := fmt.Sprintf("Usage: %s [%s] [skill name] [count?]\n\n", c.preCommand+" pts", "charname")
	helpStr += fmt.Sprintf("Available %ss:\n", "charname")

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return helpStr, nil
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		return helpStr, nil
	}

	characters := bUser.GetCharacters()
	charNames := make([]string, 0, len(characters))

	for _, char := range characters {
		charNames = append(charNames, char.GetName())
	}

	sort.Strings(charNames)
	helpStr += fmt.Sprintf("  %s", strings.Join(charNames, "\n  "))

	return helpStr, nil
}

func (c *gotCommands) helpCharsItems(user, guild, args string) (string, error) {
	helpStr := fmt.Sprintf("Usage: %s [%s] [item name] [count?]\n\n", c.preCommand+" pts", "charname")
	helpStr += fmt.Sprintf("Available %ss:\n", "charname")

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return helpStr, nil
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		return helpStr, nil
	}

	characters := bUser.GetCharacters()
	charNames := make([]string, 0, len(characters))

	for _, char := range characters {
		charNames = append(charNames, char.GetName())
	}

	sort.Strings(charNames)
	helpStr += fmt.Sprintf("  %s", strings.Join(charNames, "\n  "))

	return helpStr, nil
}

func (c *gotCommands) points(user, guild, args string) (string, error) {
	t, err := c.deps.UserAPI().NewTransaction(true)
	if err != nil {
		return "", err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		bUser, err = t.AddUser(user)
		if err != nil {
			return "", errors.Wrap(err, "could not create user")
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
		ch.SetHandler(char.GetName(), &gotPointsHandler{charName: char.GetName(), user: bUser})
	}
	retStr, err := ch.HandleLine(user, guild, args)

	if err != nil {
		return retStr, err
	}

	err = t.SaveUser(bUser)
	if err != nil {
		return "", errors.Wrap(err, "could not save points gotten")
	}

	err = t.Commit()
	if err != nil {
		return "", errors.Wrap(err, "could not save points gotten")
	}

	return retStr, nil
}

func (c *gotCommands) item(user, guild, args string) (string, error) {
	t, err := c.deps.UserAPI().NewTransaction(true)
	if err != nil {
		return "", err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		bUser, err = t.AddUser(user)
		if err != nil {
			return "", errors.Wrap(err, "could not create user")
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
		ch.SetHandler(char.GetName(), &gotItemHandler{charName: char.GetName(), user: bUser})
	}
	retStr, err := ch.HandleLine(user, guild, args)

	if err != nil {
		return retStr, err
	}

	err = t.SaveUser(bUser)
	if err != nil {
		return "", errors.Wrap(err, "could not save item gotten")
	}

	err = t.Commit()
	if err != nil {
		return "", errors.Wrap(err, "could not save item gotten")
	}

	return retStr, nil
}

// GotCommandHandler TODOC
func GotCommandHandler(deps dependencies, preCommand string) *cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: " ",
	})
	gc := gotCommands{
		preCommand: preCommand,
		deps:       deps,
	}
	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  preCommand,
		Placeholder: "type",
	})
	ch.SetHandler("pts", cmdhandler.NewLineHandler(gc.points))
	ch.SetHandler("item", cmdhandler.NewLineHandler(gc.item))

	return ch
}
