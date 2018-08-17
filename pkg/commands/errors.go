package commands

import "github.com/pkg/errors"

// ErrCharacterNameRequired is the error returned when a character name is missing
var ErrCharacterNameRequired = errors.New("character name required")

// ErrCharacterExists is the error returned when a character already exists
var ErrCharacterExists = errors.New("character already exists")

// ErrItemNameRequired is the error returned when an item name is required
var ErrItemNameRequired = errors.New("item name required")

// ErrSkillNameRequired is the error returned when a skill name is required
var ErrSkillNameRequired = errors.New("skill name required")

// ErrPositiveValueRequired is the error returned when a positive value is required
var ErrPositiveValueRequired = errors.New("positive value required")
