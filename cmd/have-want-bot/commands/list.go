package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/discord-bot-lib/util"
	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
)

type listCommands struct {
	preCommand string
	deps       dependencies
}

func (c *listCommands) items(user, guild, args string) (string, error) {
	charName := strings.TrimSpace(args)

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return "", err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(user) // add or get empty (don't save)
	if err != nil {
		return "", errors.Wrap(err, "unable to find user")
	}

	if charName != "" {
		char, err := bUser.GetCharacter(charName)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("__**%s needed items:**__\n```\n  %s\n```\n", char.GetName(), itemsDescription(char, "  ")), nil
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

	itemDescrip := "__**All needed items:**__\n```\n"
	for _, itemName := range itemNames {
		ct := itemCounts[itemName]
		itemDescrip += fmt.Sprintf("  %s x%d\n", itemName, ct)
	}
	itemDescrip += "```\n"
	return itemDescrip, nil
}

func (c *listCommands) points(user, guild, args string) (string, error) {
	charName := strings.TrimSpace(args)

	t, err := c.deps.UserAPI().NewTransaction(false)
	if err != nil {
		return "", err
	}
	defer util.CheckDefer(t.Rollback)

	bUser, err := t.AddUser(user) // add or get empty (don't save)
	if err != nil {
		return "", errors.Wrap(err, "unable to find user")
	}

	if charName != "" {
		char, err := bUser.GetCharacter(charName)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("__**%s needed points:**__\n```\n  %s\n```\n", char.GetName(), skillsDescription(char, "  ")), nil
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

	skillDescrip := "__**All needed skills:**__\n```\n"
	for _, skillName := range skillNames {
		ct := skillCounts[skillName]
		skillDescrip += fmt.Sprintf("  %s x%d\n", skillName, ct)
	}
	skillDescrip += "```\n"
	return skillDescrip, nil
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
