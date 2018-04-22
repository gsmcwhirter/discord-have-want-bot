package errors

import "errors"

var ErrCharacterNameRequired = errors.New("character name required")
var ErrCharacterExists = errors.New("character already exists")
var ErrItemNameRequired = errors.New("item name required")
var ErrSkillNameRequired = errors.New("skill name required")
var ErrPositiveValueRequired = errors.New("positive value required")
