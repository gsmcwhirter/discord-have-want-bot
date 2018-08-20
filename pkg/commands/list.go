package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/go-util/deferutil"
	"github.com/gsmcwhirter/go-util/parser"
)

type listCommands struct {
	preCommand string
	deps       dependencies
}

func (c *listCommands) items(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	charName := strings.TrimSpace(msg.Contents())

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, err
	}
	defer deferutil.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(msg.UserID().ToString()) // add or get empty (don't save)
	if err != nil {
		return r, errors.Wrap(err, "unable to find user")
	}

	if charName != "" {
		char, err := bUser.GetCharacter(charName)
		if err != nil {
			return r, err
		}

		itemDescrip, itemCt := itemsDescription(char, "")

		r.Title = fmt.Sprintf("__%s__", char.GetName())
		r.Description = "Remember, you can call `need item [charname] [item]` and `got item [charname] [item]` to add and remove items to this list."
		r.Fields = []cmdhandler.EmbedField{
			{
				Name: fmt.Sprintf("*Needed Items (%d)*", itemCt),
				Val:  fmt.Sprintf("```\n%s\n```\n", itemDescrip),
			},
		}

		return r, nil
	}

	var total uint64
	itemCounts := map[string]uint64{}
	for _, char := range bUser.GetCharacters() {
		for _, item := range char.GetNeededItems() {
			itemCounts[item.Name()] += item.Count()
			total += item.Count()
		}
	}

	itemNames := make([]string, 0, len(itemCounts))
	for itemName := range itemCounts {
		itemNames = append(itemNames, itemName)
	}
	sort.Strings(itemNames)

	itemDescrip := ""
	for _, itemName := range itemNames {
		ct := itemCounts[itemName]
		itemDescrip += fmt.Sprintf("%s x%d\n", itemName, ct)
	}

	r.Title = "__All Characters__"
	r.Description = "Remember, you can call `need item [charname] [item]` and `got item [charname] [item]` to add and remove items to this list."
	r.Fields = []cmdhandler.EmbedField{
		{
			Name: fmt.Sprintf("*Needed Items (%d)*", total),
			Val:  fmt.Sprintf("```\n%s\n```\n", itemDescrip),
		},
	}

	return r, nil
}

func (c *listCommands) points(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	charName := strings.TrimSpace(msg.Contents())

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, err
	}
	defer deferutil.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(msg.UserID().ToString()) // add or get empty (don't save)
	if err != nil {
		return r, errors.Wrap(err, "unable to find user")
	}

	if charName != "" {
		char, err := bUser.GetCharacter(charName)
		if err != nil {
			return r, err
		}

		skillsDescrip, skillsCt := skillsDescription(char, "  ")

		r.Title = fmt.Sprintf("__%s__", char.GetName())
		r.Description = "Remember, you can call `need pts [charname] [item]` and `got pts [charname] [item]` to add and remove items to this list."
		r.Fields = []cmdhandler.EmbedField{
			{
				Name: fmt.Sprintf("*Needed Points (%d)*", skillsCt),
				Val:  fmt.Sprintf("```\n%s\n```\n", skillsDescrip),
			},
		}

		return r, nil
	}

	var total uint64
	skillCounts := map[string]uint64{}
	for _, char := range bUser.GetCharacters() {
		for _, skill := range char.GetNeededSkills() {
			skillCounts[skill.Name()] += skill.Points()
			total += skill.Points()
		}
	}

	skillNames := make([]string, 0, len(skillCounts))
	for k := range skillCounts {
		skillNames = append(skillNames, k)
	}

	sort.Strings(skillNames)

	skillDescrip := ""
	for _, skillName := range skillNames {
		ct := skillCounts[skillName]
		skillDescrip += fmt.Sprintf("%s x%d\n", skillName, ct)
	}

	r.Title = "__All Characters__"
	r.Description = "Remember, you can call `need pts [charname] [item]` and `got pts [charname] [item]` to add and remove items to this list."
	r.Fields = []cmdhandler.EmbedField{
		{
			Name: fmt.Sprintf("*Needed Points (%d)*", total),
			Val:  fmt.Sprintf("```\n%s\n```\n", skillDescrip),
		},
	}

	return r, nil
}

func (c *listCommands) transmutes(msg cmdhandler.Message) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: cmdhandler.UserMentionString(msg.UserID()),
	}

	charName := strings.TrimSpace(msg.Contents())

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, err
	}
	defer deferutil.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(msg.UserID().ToString()) // add or get empty (don't save)
	if err != nil {
		return r, errors.Wrap(err, "unable to find user")
	}

	if charName != "" {
		char, err := bUser.GetCharacter(charName)
		if err != nil {
			return r, err
		}

		transDescrip, transCt := transDescription(char, "  ")

		r.Title = fmt.Sprintf("__%s__", char.GetName())
		r.Description = "Remember, you can call `need trans [charname] [item]` and `got trans [charname] [item]` to add and remove items to this list."
		r.Fields = []cmdhandler.EmbedField{
			{
				Name: fmt.Sprintf("*Needed Transmutes (%d; %d stones)*", transCt, transCt*50),
				Val:  fmt.Sprintf("```\n%s\n```\n", transDescrip),
			},
		}

		return r, nil
	}

	var total uint64
	itemCounts := map[string]uint64{}
	for _, char := range bUser.GetCharacters() {
		for _, trans := range char.GetNeededTransmutes() {
			itemCounts[trans.Name()] += trans.Count()
			total += trans.Count()
		}
	}

	itemNames := make([]string, 0, len(itemCounts))
	for k := range itemCounts {
		itemNames = append(itemNames, k)
	}

	sort.Strings(itemNames)

	itemDescrip := ""
	for _, itemName := range itemNames {
		ct := itemCounts[itemName]
		itemDescrip += fmt.Sprintf("%s x%d\n", itemName, ct)
	}

	r.Title = "__All Characters__"
	r.Description = "Remember, you can call `need trans [charname] [item]` and `got trans [charname] [item]` to add and remove items to this list."
	r.Fields = []cmdhandler.EmbedField{
		{
			Name: fmt.Sprintf("*Needed Transmutes (%d; %d stones)*", total, total*50),
			Val:  fmt.Sprintf("```\n%s\n```\n", itemDescrip),
		},
	}

	return r, nil
}

// ListCommandHandler creates a command handler for !list commands
func ListCommandHandler(deps dependencies, preCommand string) (*cmdhandler.CommandHandler, error) {
	p := parser.NewParser(parser.Options{
		CmdIndicator: " ",
	})
	cc := listCommands{
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

	ch.SetHandler("items", cmdhandler.NewMessageHandler(cc.items))
	ch.SetHandler("pts", cmdhandler.NewMessageHandler(cc.points))
	ch.SetHandler("trans", cmdhandler.NewMessageHandler(cc.transmutes))

	return ch, nil
}
