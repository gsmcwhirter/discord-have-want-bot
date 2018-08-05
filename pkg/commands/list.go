package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/discord-bot-lib/cmdhandler"
	"github.com/gsmcwhirter/discord-bot-lib/util"
	"github.com/gsmcwhirter/go-util/parser"
)

type listCommands struct {
	preCommand string
	deps       dependencies
}

func (c *listCommands) items(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: user,
	}

	charName := strings.TrimSpace(args)

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(user) // add or get empty (don't save)
	if err != nil {
		return r, errors.Wrap(err, "unable to find user")
	}

	if charName != "" {
		char, err := bUser.GetCharacter(charName)
		if err != nil {
			return r, err
		}

		r.Title = fmt.Sprintf("__%s__", char.GetName())
		r.Description = "Remember, you can call `need item [charname] [item]` and `got item [charname] [item]` to add and remove items to this list."
		r.Fields = []cmdhandler.EmbedField{
			{
				Name: "*Needed Items*",
				Val:  fmt.Sprintf("```\n%s\n```\n", itemsDescription(char, "")),
			},
		}

		return r, nil
	}

	itemCounts := map[string]uint64{}
	for _, char := range bUser.GetCharacters() {
		for _, item := range char.GetNeededItems() {
			itemCounts[item.Name()] += item.Count()
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
			Name: "*Needed Items*",
			Val:  fmt.Sprintf("```\n%s\n```\n", itemDescrip),
		},
	}

	return r, nil
}

func (c *listCommands) points(user, guild, args string) (cmdhandler.Response, error) {
	r := &cmdhandler.EmbedResponse{
		To: user,
	}

	charName := strings.TrimSpace(args)

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return r, err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(user) // add or get empty (don't save)
	if err != nil {
		return r, errors.Wrap(err, "unable to find user")
	}

	if charName != "" {
		char, err := bUser.GetCharacter(charName)
		if err != nil {
			return r, err
		}

		r.Title = fmt.Sprintf("__%s__", char.GetName())
		r.Description = "Remember, you can call `need pts [charname] [item]` and `got pts [charname] [item]` to add and remove items to this list."
		r.Fields = []cmdhandler.EmbedField{
			{
				Name: "*Needed Points*",
				Val:  fmt.Sprintf("```\n%s\n```\n", skillsDescription(char, "  ")),
			},
		}

		return r, nil
	}

	skillCounts := map[string]uint64{}
	for _, char := range bUser.GetCharacters() {
		for _, skill := range char.GetNeededSkills() {
			skillCounts[skill.Name()] += skill.Points()
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
			Name: "*Needed Points(",
			Val:  fmt.Sprintf("```\n%s\n```\n", skillDescrip),
		},
	}

	return r, nil
}

// ListCommandHandler TODOC
func ListCommandHandler(deps dependencies, preCommand string) *cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: " ",
	})
	cc := listCommands{
		preCommand: preCommand,
		deps:       deps,
	}
	ch := cmdhandler.NewCommandHandler(p, cmdhandler.Options{
		PreCommand:  preCommand,
		Placeholder: "type",
	})
	ch.SetHandler("items", cmdhandler.NewLineHandler(cc.items))
	ch.SetHandler("points", cmdhandler.NewLineHandler(cc.points))

	return ch
}
