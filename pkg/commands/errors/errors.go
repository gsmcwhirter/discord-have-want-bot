package errors

import "errors"

// ErrCharacterNameRequired TODOC
var ErrCharacterNameRequired = errors.New("character name required")

// ErrCharacterExists TODOC
var ErrCharacterExists = errors.New("character already exists")

// ErrItemNameRequired TODOC
var ErrItemNameRequired = errors.New("item name required")

// ErrSkillNameRequired TODOC
var ErrSkillNameRequired = errors.New("skill name required")

// ErrPositiveValueRequired TODOC
var ErrPositiveValueRequired = errors.New("positive value required")
