package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/go-util/deferutil"
	"github.com/gsmcwhirter/go-util/parser"

	"github.com/gsmcwhirter/discord-have-want-bot/pkg/storage"
)

type charCommands struct {
	preCommand string
	deps       dependencies
}

func (c *charCommands) show(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	charName := strings.TrimSpace(msg.Contents())

	if len(charName) == 0 {
		return r, ErrCharacterNameRequired
	}

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, err
	}
	defer deferutil.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(msg.UserID().ToString()) // add or get empty (don't save)
	if err != nil {
		return r, errors.Wrap(err, "unable to find user")
	}

	char, err := bUser.GetCharacter(charName)
	if err != nil {
		return r, err
	}

	itemDescrip, itemCt := itemsDescription(char, "")
	skillDescrip, skillCt := skillsDescription(char, "")
	transDescrip, transCt := transDescription(char, "")

	r.Title = fmt.Sprintf("__%s__", char.GetName())
	r.Description = "Remember, you can call `need item [charname] [item]` and `need pts [charname] [item]` to add items to these lists. You can also call `got item [charname] [item]` and `got pts [charname] [item]` to remove items from this list."
	r.Fields = []cmdhandler.EmbedField{
		{
			Name: fmt.Sprintf("*Needed Items (%d)*", itemCt),
			Val:  fmt.Sprintf("```\n%s\n```\n", itemDescrip),
		},
		{
			Name: fmt.Sprintf("*Needed Transmutes (%d; %d stones)*", transCt, transCt*50),
			Val:  fmt.Sprintf("```\n%s\n```\n", transDescrip),
		},
		{
			Name: fmt.Sprintf("*Needed Skills (%d)*", skillCt),
			Val:  fmt.Sprintf("```\n%s\n```\n", skillDescrip),
		},
	}

	return r, nil
}

func (c *charCommands) create(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleEmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	charName := strings.TrimSpace(msg.Contents())

	if len(charName) == 0 {
		return r, ErrCharacterNameRequired
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

	r.Description = "character created"
	return r, nil
}

func (c *charCommands) delete(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.SimpleEmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	charName := strings.TrimSpace(msg.Contents())

	if len(charName) == 0 {
		return r, ErrCharacterNameRequired
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

	r.Description = "character deleted"
	return r, nil
}

func (c *charCommands) list(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, err
	}
	defer deferutil.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(msg.UserID().ToString()) // add or get empty (don't save)
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

func (c *charCommands) help(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	r.Description = fmt.Sprintf("Usage: %s [%s]\n\n", c.preCommand, "action")
	r.Fields = []cmdhandler.EmbedField{
		{
			Name: "*Available Actions*",
			Val:  "- help\n- list\n- show [charname]\n- create [charname]\n- delete [charname]\n",
		},
	}

	return r, nil
}

// CharCommandHandler creates a command handler for !char commands
func CharCommandHandler(deps dependencies, preCommand string) (*cmdhandler.CommandHandler, error) {
	p := parser.NewParser(parser.Options{
		CmdIndicator: " ",
	})
	cc := charCommands{
		preCommand: preCommand,
		deps:       deps,
	}
	ch, err := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  preCommand,
		Placeholder: "action",
	})
	if err != nil {
		return nil, err
	}

	ch.SetHandler("", cmdhandler.NewMessageHandler(cc.help))
	ch.SetHandler("help", cmdhandler.NewMessageHandler(cc.help))
	ch.SetHandler("list", cmdhandler.NewMessageHandler(cc.list))
	ch.SetHandler("show", cmdhandler.NewMessageHandler(cc.show))
	ch.SetHandler("create", cmdhandler.NewMessageHandler(cc.create))
	ch.SetHandler("delete", cmdhandler.NewMessageHandler(cc.delete))

	return ch, nil
}
