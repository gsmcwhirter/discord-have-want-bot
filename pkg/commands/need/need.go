package need

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gsmcwhirter/eso-discord/pkg/cmdhandler"
	cmderrors "github.com/gsmcwhirter/eso-discord/pkg/commands/errors"
	"github.com/gsmcwhirter/eso-discord/pkg/parser"
	"github.com/gsmcwhirter/eso-discord/pkg/storage"
	errorsx "github.com/pkg/errors"
)

type charItemHandler struct {
	user     storage.User
	charName string
}

func (h charItemHandler) HandleLine(user string, args []rune) (string, error) {
	args, ctRunes := parser.MaybeCount(args)

	itemName := strings.TrimSpace(string(args))

	if len(itemName) == 0 {
		return "", cmderrors.ErrItemNameRequired
	}

	ctStr := strings.TrimSpace(string(ctRunes))
	if ctStr == "" {
		ctStr = "1"
	}

	ct, err := strconv.Atoi(ctStr)
	if err != nil {
		return "", errorsx.Wrap(err, "could not interpret count to adjust item needs")
	}

	if ct < 0 {
		return "", cmderrors.ErrPositiveValueRequired
	}

	char, err := h.user.GetCharacter(h.charName)
	if err != nil {
		return "", errorsx.Wrap(err, "could not find character to adjust item needs")
	}

	char.IncrNeededItem(itemName, uint64(ct))

	return fmt.Sprintf("marked %s as needing +%d of %s", h.charName, ct, itemName), nil
}

type charPointsHandler struct {
	user     storage.User
	charName string
}

func (h charPointsHandler) HandleLine(user string, args []rune) (string, error) {
	args, ctRunes := parser.MaybeCount(args)

	skillName := strings.TrimSpace(string(args))

	if len(skillName) == 0 {
		return "", cmderrors.ErrSkillNameRequired
	}

	ctStr := strings.TrimSpace(string(ctRunes))
	if ctStr == "" {
		ctStr = "1"
	}

	ct, err := strconv.Atoi(ctStr)
	if err != nil {
		return "", errorsx.Wrap(err, "could not interpret count to adjust skill needs")
	}

	if ct < 0 {
		return "", cmderrors.ErrPositiveValueRequired
	}

	char, err := h.user.GetCharacter(h.charName)
	if err != nil {
		return "", errorsx.Wrap(err, "could not find character to adjust skill needs")
	}

	char.IncrNeededSkill(skillName, uint64(ct))

	return fmt.Sprintf("marked %s as needing +%d points in %s", h.charName, ct, skillName), nil
}

type needCommands struct {
	userAPI storage.UserAPI
}

func (c needCommands) points(user string, args []rune) (string, error) {
	t, err := c.userAPI.NewTransaction(true)
	if err != nil {
		return "", err
	}
	defer t.Rollback()

	bUser, err := t.GetUser(user)
	if err != nil {
		bUser, err = t.AddUser(user)
		if err != nil {
			return "", errorsx.Wrap(err, "could not create user")
		}
	}

	characters := bUser.GetCharacters()
	charNames := make([]string, len(characters))
	for i, char := range characters {
		charNames[i] = char.GetName()
	}

	p := parser.NewParser(parser.Options{
		CmdIndicator:  ' ',
		KnownCommands: charNames,
	})
	ch := cmdhandler.NewCommandHandler(p)
	for _, char := range characters {
		ch.SetHandler(char.GetName(), charPointsHandler{charName: char.GetName(), user: bUser})
	}
	retStr, err := ch.HandleLine(user, args)

	if err != nil {
		return retStr, err
	}

	err = t.SaveUser(bUser)
	if err != nil {
		return "", errorsx.Wrap(err, "could not save points need")
	}

	err = t.Commit()
	if err != nil {
		return "", errorsx.Wrap(err, "could not save points need")
	}

	return retStr, nil
}

func (c needCommands) item(user string, args []rune) (string, error) {
	t, err := c.userAPI.NewTransaction(true)
	if err != nil {
		return "", err
	}
	defer t.Rollback()

	bUser, err := t.GetUser(user)
	if err != nil {
		bUser, err = t.AddUser(user)
		if err != nil {
			return "", errorsx.Wrap(err, "could not create user")
		}
	}

	characters := bUser.GetCharacters()
	charNames := make([]string, len(characters))
	for i, char := range characters {
		charNames[i] = char.GetName()
	}

	p := parser.NewParser(parser.Options{
		CmdIndicator:  ' ',
		KnownCommands: charNames,
	})
	ch := cmdhandler.NewCommandHandler(p)
	for _, char := range characters {
		ch.SetHandler(char.GetName(), charItemHandler{charName: char.GetName(), user: bUser})
	}
	retStr, err := ch.HandleLine(user, args)

	if err != nil {
		return retStr, err
	}

	err = t.SaveUser(bUser)
	if err != nil {
		return "", errorsx.Wrap(err, "could not save item need")
	}

	err = t.Commit()
	if err != nil {
		return "", errorsx.Wrap(err, "could not save item need")
	}

	return retStr, nil
}

// CommandHandler TODOC
func CommandHandler(userAPI storage.UserAPI) cmdhandler.CommandHandler {
	p := parser.NewParser(parser.Options{
		CmdIndicator: ' ',
		KnownCommands: []string{
			"help",
			"pts",
			"item",
		},
	})
	nc := needCommands{
		userAPI: userAPI,
	}
	ch := cmdhandler.NewCommandHandler(p)
	ch.SetHandler("pts", cmdhandler.NewLineHandler(nc.points))
	ch.SetHandler("item", cmdhandler.NewLineHandler(nc.item))

	return ch
}
