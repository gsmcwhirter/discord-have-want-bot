package commands

import (
	"fmt"
	"strings"

	"github.com/gsmcwhirter/eso-discord/pkg/storage"
)

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
