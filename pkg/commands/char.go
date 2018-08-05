package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/discord-bot-lib/util"
	"github.com/gsmcwhirter/go-util/parser"

	"github.com/gsmcwhirter/discord-have-want-bot/pkg/storage"
)

type charCommands struct {
	preCommand string
	deps       dependencies
}

func (c *charCommands) show(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: user,
	}

	charName := strings.TrimSpace(args)

	if len(charName) == 0 {
		return r, ErrCharacterNameRequired
	}

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(user) // add or get empty (don't save)
	if err != nil {
		return r, errors.Wrap(err, "unable to find user")
	}

	char, err := bUser.GetCharacter(charName)
	if err != nil {
		return r, err
	}

	r.Title = fmt.Sprintf("__%s__", char.GetName())
	r.Description = "Remember, you can call `need item [charname] [item]` and `need pts [charname] [item]` to add items to these lists. You can also call `got item [charname] [item]` and `got pts [charname] [item]` to remove items from this list."
	r.Fields = []cmdhandler.EmbedField{
		{
			Name: "*Needed Items*",
			Val:  fmt.Sprintf("```\n%s\n```\n", itemsDescription(char, "")),
		},
		{
			Name: "*Needed Skills*",
			Val:  fmt.Sprintf("```\n%s\n```\n", skillsDescription(char, "")),
		},
	}

	// 	charDescription := fmt.Sprintf(`__**%[2]s**__

	//   Needed Items:
	// %[1]s
	// 	%[3]s
	// %[1]s

	//   Needed Skills:
	// %[1]s
	// 	%[4]s
	// %[1]s
	// `, "```", char.GetName(), itemsDescription(char, "    "), skillsDescription(char, "    "))

	return r, nil
}

func (c *charCommands) create(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleResponse{
		To: user,
	}

	charName := strings.TrimSpace(args)

	if len(charName) == 0 {
		return r, ErrCharacterNameRequired
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
			return r, errors.Wrap(err, "could not create character")
		}
	}

	_, err = bUser.GetCharacter(charName)
	if err != storage.ErrCharacterNotExist {
		if err != nil {
			return r, errors.Wrap(err, "could not verify character does not exist")
		}

		return r, ErrCharacterExists
	}

	_ = bUser.AddCharacter(charName)
	err = t.SaveUser(bUser)
	if err != nil {
		return r, errors.Wrap(err, "could not save new character")
	}

	err = t.Commit()
	if err != nil {
		return r, errors.Wrap(err, "could not save new character")
	}

	r.Content = "character created"
	return r, nil
}

func (c *charCommands) delete(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleResponse{
		To: user,
	}

	charName := strings.TrimSpace(args)

	if len(charName) == 0 {
		return r, ErrCharacterNameRequired
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
			return r, errors.Wrap(err, "could not create character")
		}
	}

	_, err = bUser.GetCharacter(charName)
	if err != nil {
		return r, errors.Wrap(err, "could not find character")
	}

	bUser.DeleteCharacter(charName)
	err = t.SaveUser(bUser)
	if err != nil {
		return r, errors.Wrap(err, "could not delete character")
	}

	err = t.Commit()
	if err != nil {
		return r, errors.Wrap(err, "could not delete character")
	}

	r.Content = "character deleted"
	return r, nil
}

func (c *charCommands) list(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: user,
	}

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(user) // add or get empty (don't save)
	if err != nil {
		return r, errors.Wrap(err, "unable to find user")
	}

	chars := bUser.GetCharacters()
	charNames := make([]string, 0, len(chars))
	for _, char := range chars {
		charNames = append(charNames, char.GetName())
	}
	sort.Strings(charNames)

	r.Title = "__Character List__"
	r.Description = "Remember, you can call `char create [charname]` and `char delete [charname]` to edit this list."
	r.Fields = []cmdhandler.EmbedField{
		{
			Name: "*Your Characters*",
			Val:  fmt.Sprintf("```\n%s\n```\n", strings.Join(charNames, "\n")),
		},
	}

	return r, nil
}

func (c *charCommands) help(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleResponse{
		To: user,
	}

	r.Content = fmt.Sprintf("Usage: %s [%s]\n\n", c.preCommand, "action")
	r.Content += "Available actions:\n  help\n  list\n  show [charname]\n  create [charname]\n  delete [charname]\n"

	return r, nil
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
