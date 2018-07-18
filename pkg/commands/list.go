package commands

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
	"github.com/gsmcwhirter/eso-discord/pkg/util"
)

type listCommands struct {
	preCommand string
	deps       dependencies
}

func (c listCommands) items(user string, args []rune) (string, error) {
	charName := strings.TrimSpace(string(args))

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

	itemDescrip := "__**All needed items:**__\n```\n"
	for itemName, ct := range itemCounts {
		itemDescrip += fmt.Sprintf("  %s x%d\n", itemName, ct)
	}
	itemDescrip += "```\n"
	return itemDescrip, nil
}

func (c listCommands) points(user string, args []rune) (string, error) {
	charName := strings.TrimSpace(string(args))

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

	skillDescrip := "__**All needed skills:**__\n```\n"
	for skillName, ct := range skillCounts {
		skillDescrip += fmt.Sprintf("  %s x%d\n", skillName, ct)
	}
	skillDescrip += "```\n"
	return skillDescrip, nil
}

// ListCommandHandler TODOC
func ListCommandHandler(deps dependencies, preCommand string) *cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: ' ',
		KnownCommands: []string{
			"help",
			"items",
			"points",
		},
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
