package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gsmcwhirter/discord-have-want-bot/pkg/storage"
)

func skillsDescription(char storage.Character, indent string) (string, uint64) {
	var total uint64
	skills := char.GetNeededSkills()
	skillStrings := make([]string, len(skills))
	for i, skill := range skills {
		skillStrings[i] = fmt.Sprintf("%s x%d", skill.Name(), skill.Points())
		total += skill.Points()
	}

	sort.Strings(skillStrings)

	return strings.Join(skillStrings, fmt.Sprintf("\n%s", indent)), total
}

func itemsDescription(char storage.Character, indent string) (string, uint64) {
	var total uint64
	items := char.GetNeededItems()
	itemStrings := make([]string, len(items))
	for i, item := range items {
		itemStrings[i] = fmt.Sprintf("%s x%d", item.Name(), item.Count())
		total += item.Count()
	}

	sort.Strings(itemStrings)

	return strings.Join(itemStrings, fmt.Sprintf("\n%s", indent)), total
}

func transDescription(char storage.Character, indent string) (string, uint64) {
	var total uint64
	items := char.GetNeededTransmutes()
	itemStrings := make([]string, len(items))
	for i, item := range items {
		itemStrings[i] = fmt.Sprintf("%s x%d", item.Name(), item.Count())
		total += item.Count()
	}

	sort.Strings(itemStrings)

	return strings.Join(itemStrings, fmt.Sprintf("\n%s", indent)), total
}
