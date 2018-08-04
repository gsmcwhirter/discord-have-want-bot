package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/discord-bot-lib/util"
	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"

	"github.com/gsmcwhirter/eso-discord/pkg/storage"
)

const triplet = "```"

type charCommands struct {
	preCommand string
	deps       dependencies
}

func (c *charCommands) show(user, guild, args string) (string, error) {
	charName := strings.TrimSpace(args)

	if len(charName) == 0 {
		return "", ErrCharacterNameRequired
	}

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return "", err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(user) // add or get empty (don't save)
	if err != nil {
		return "", errors.Wrap(err, "unable to find user")
	}

	char, err := bUser.GetCharacter(charName)
	if err != nil {
		return "", err
	}

	charDescription := fmt.Sprintf(`__**%[2]s**__

  Needed Items:
%[1]s
	%[3]s
%[1]s
		
  Needed Skills:
%[1]s
	%[4]s
%[1]s
`, triplet, char.GetName(), itemsDescription(char, "    "), skillsDescription(char, "    "))

	return charDescription, nil
}

func (c *charCommands) create(user, guild, args string) (string, error) {
	charName := strings.TrimSpace(args)

	if len(charName) == 0 {
		return "", ErrCharacterNameRequired
	}

	t, err := c.deps.UserAPI().NewTransaction(true)
	if err != nil {
		return "", err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		bUser, err = t.AddUser(user)
		if err != nil {
			return "", errors.Wrap(err, "could not create character")
		}
	}

	_, err = bUser.GetCharacter(charName)
	if err != storage.ErrCharacterNotExist {
		if err != nil {
			return "", errors.Wrap(err, "could not verify character does not exist")
		}

		return "", ErrCharacterExists
	}

	_ = bUser.AddCharacter(charName)
	err = t.SaveUser(bUser)
	if err != nil {
		return "", errors.Wrap(err, "could not save new character")
	}

	err = t.Commit()
	if err != nil {
		return "", errors.Wrap(err, "could not save new character")
	}

	return "character created", nil
}

func (c *charCommands) delete(user, guild, args string) (string, error) {
	charName := strings.TrimSpace(args)

	if len(charName) == 0 {
		return "", ErrCharacterNameRequired
	}

	t, err := c.deps.UserAPI().NewTransaction(true)
	if err != nil {
		return "", err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.GetUser(user)
	if err != nil {
		bUser, err = t.AddUser(user)
		if err != nil {
			return "", errors.Wrap(err, "could not create character")
		}
	}

	_, err = bUser.GetCharacter(charName)
	if err != nil {
		return "", errors.Wrap(err, "could not find character")
	}

	bUser.DeleteCharacter(charName)
	err = t.SaveUser(bUser)
	if err != nil {
		return "", errors.Wrap(err, "could not delete character")
	}

	err = t.Commit()
	if err != nil {
		return "", errors.Wrap(err, "could not delete character")
	}

	return "character deleted", nil
}

func (c *charCommands) list(user, guild, args string) (string, error) {
	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return "", err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(user) // add or get empty (don't save)
	if err != nil {
		return "", errors.Wrap(err, "unable to find user")
	}

	chars := bUser.GetCharacters()
	charNames := make([]string, 0, len(chars))
	for _, char := range chars {
		charNames = append(charNames, char.GetName())
	}
	sort.Strings(charNames)

	charList := "__**Your characters:**__\n```\n"
	for _, charName := range charNames {
		charList += fmt.Sprintf("  %s\n", charName)
	}
	charList += "```\n"
	return charList, nil
}

func (c *charCommands) help(user, guild, args string) (string, error) {
	helpStr := fmt.Sprintf("Usage: %s [%s]\n\n", c.preCommand, "action")
	helpStr += "Available actions:\n  help\n  list\n  show [charname]\n  create [charname]\n  delete [charname]\n"
	return helpStr, nil
}

// CharCommandHandler TODOC
func CharCommandHandler(deps dependencies, preCommand string) *cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: " ",
	})
	cc := charCommands{
		preCommand: preCommand,
		deps:       deps,
	}
	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  preCommand,
		Placeholder: "action",
	})
	ch.SetHandler("", cmdhandler.NewLineHandler(cc.help))
	ch.SetHandler("help", cmdhandler.NewLineHandler(cc.help))
	ch.SetHandler("list", cmdhandler.NewLineHandler(cc.list))
	ch.SetHandler("show", cmdhandler.NewLineHandler(cc.show))
	ch.SetHandler("create", cmdhandler.NewLineHandler(cc.create))
	ch.SetHandler("delete", cmdhandler.NewLineHandler(cc.delete))

	return ch
}
