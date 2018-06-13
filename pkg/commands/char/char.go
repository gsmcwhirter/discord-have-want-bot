package char

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	cmderrors "github.com/gsmcwhirter/eso-discord/pkg/commands/errors"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
	"github.com/gsmcwhirter/eso-discord/pkg/storage"
	"github.com/gsmcwhirter/eso-discord/pkg/util"
)

type dependencies interface {
	UserAPI() storage.UserAPI
}

func skillsDescription(char storage.Character, indent string) string {
	skills := char.GetNeededSkills()
	skillStrings := make([]string, len(skills))
	for i, skill := range skills {
		skillStrings[i] = fmt.Sprintf("%s x%d", skill.Name(), skill.Points())
	}
	return strings.Join(skillStrings, fmt.Sprintf("\n%s", indent))
}

func itemsDescription(char storage.Character, indent string) string {
	items := char.GetNeededItems()
	itemStrings := make([]string, len(items))
	for i, item := range items {
		itemStrings[i] = fmt.Sprintf("%s x%d", item.Name(), item.Count())
	}
	return strings.Join(itemStrings, fmt.Sprintf("\n%s", indent))
}

type charCommands struct {
	deps dependencies
}

func (c charCommands) show(user string, args []rune) (string, error) {
	charName := strings.TrimSpace(string(args))

	if len(charName) == 0 {
		return "", cmderrors.ErrCharacterNameRequired
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

	charDescription := fmt.Sprintf(`%s

  Needed Items:
    %s
		
  Needed Skills
    %s
`, char.GetName(), itemsDescription(char, "    "), skillsDescription(char, "    "))

	return charDescription, nil
}

func (c charCommands) items(user string, args []rune) (string, error) {
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

		return fmt.Sprintf("%s needed items:\n  %s", char.GetName(), itemsDescription(char, "  ")), nil
	}

	itemCounts := map[string]uint64{}
	for _, char := range bUser.GetCharacters() {
		for _, item := range char.GetNeededItems() {
			itemCounts[item.Name()] += item.Count()
		}
	}

	itemDescrip := "All needed items:\n"
	for itemName, ct := range itemCounts {
		itemDescrip += fmt.Sprintf("  %s x%d\n", itemName, ct)
	}
	return itemDescrip, nil
}

func (c charCommands) points(user string, args []rune) (string, error) {
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

		return fmt.Sprintf("%s needed points:\n  %s", char.GetName(), skillsDescription(char, "  ")), nil
	}

	skillCounts := map[string]uint64{}
	for _, char := range bUser.GetCharacters() {
		for _, skill := range char.GetNeededSkills() {
			skillCounts[skill.Name()] += skill.Points()
		}
	}

	skillDescrip := "All needed skills:\n"
	for skillName, ct := range skillCounts {
		skillDescrip += fmt.Sprintf("  %s x%d\n", skillName, ct)
	}
	return skillDescrip, nil
}

func (c charCommands) create(user string, args []rune) (string, error) {
	charName := strings.TrimSpace(string(args))

	if len(charName) == 0 {
		return "", cmderrors.ErrCharacterNameRequired
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

		return "", cmderrors.ErrCharacterExists
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

func (c charCommands) delete(user string, args []rune) (string, error) {
	charName := strings.TrimSpace(string(args))

	if len(charName) == 0 {
		return "", cmderrors.ErrCharacterNameRequired
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

func (c charCommands) blank(user string, args []rune) (string, error) {
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
	charList := "Your characters:\n"
	for _, char := range chars {
		charList += fmt.Sprintf("  %s\n", char.GetName())
	}
	return charList, nil
}

// CommandHandler TODOC
func CommandHandler(deps dependencies) cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: ' ',
		KnownCommands: []string{
			"",
			"help",
			"create",
			"delete",
			"show",
			"items",
			"points",
		},
	})
	cc := charCommands{
		deps: deps,
	}
	ch := cmdhandler.NewCommandHandler(p)
	ch.SetHandler("", cmdhandler.NewLineHandler(cc.blank))
	ch.SetHandler("show", cmdhandler.NewLineHandler(cc.show))
	ch.SetHandler("items", cmdhandler.NewLineHandler(cc.items))
	ch.SetHandler("points", cmdhandler.NewLineHandler(cc.points))
	ch.SetHandler("create", cmdhandler.NewLineHandler(cc.create))
	ch.SetHandler("delete", cmdhandler.NewLineHandler(cc.delete))

	return ch
}
