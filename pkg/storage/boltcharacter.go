package storage

import "errors"

// ErrSkillNotExist TODOC
var ErrSkillNotExist = errors.New("skill does not exist")

// ErrItemNotExist TODOC
var ErrItemNotExist = errors.New("item does not exist")

type boltCharacter struct {
	protoCharacter *ProtoCharacter
}

func (c *boltCharacter) GetName() string {
	return c.protoCharacter.Name
}

func (c *boltCharacter) GetNeededSkill(name string) (Skill, error) {
	if c.protoCharacter.NeededSkills == nil {
		return nil, ErrSkillNotExist
	}

	protoSkill, ok := c.protoCharacter.NeededSkills[name]
	if !ok {
		return nil, ErrSkillNotExist
	}

	return boltSkill{protoSkill}, nil
}

func (c *boltCharacter) GetNeededSkills() []Skill {
	if c.protoCharacter.NeededSkills == nil {
		return []Skill{}
	}

	skills := make([]Skill, len(c.protoCharacter.NeededSkills))
	i := 0
	for _, protoSkill := range c.protoCharacter.NeededSkills {
		skills[i] = boltSkill{protoSkill}
		i++
	}

	return skills
}

func (c *boltCharacter) GetNeededItem(name string) (Item, error) {
	if c.protoCharacter.NeededItems == nil {
		return nil, ErrItemNotExist
	}

	protoItem, ok := c.protoCharacter.NeededItems[name]
	if !ok {
		return nil, ErrItemNotExist
	}

	return boltItem{protoItem}, nil
}

func (c *boltCharacter) GetNeededItems() []Item {
	if c.protoCharacter.NeededItems == nil {
		return []Item{}
	}

	items := make([]Item, len(c.protoCharacter.NeededItems))
	i := 0
	for _, protoItem := range c.protoCharacter.NeededItems {
		items[i] = boltItem{protoItem}
		i++
	}

	return items
}

func (c *boltCharacter) SetName(name string) {
	c.protoCharacter.Name = name
}

func (c *boltCharacter) IncrNeededSkill(name string, amt uint64) {
	if c.protoCharacter.NeededSkills == nil {
		c.protoCharacter.NeededSkills = map[string]*ProtoSkill{}
	}

	s, ok := c.protoCharacter.NeededSkills[name]
	if !ok {
		c.protoCharacter.NeededSkills[name] = &ProtoSkill{Name: name, Ct: amt}
	} else {
		s.Ct += amt
	}
}

func (c *boltCharacter) DecrNeededSkill(name string, amt uint64) {
	if c.protoCharacter.NeededSkills == nil {
		return
	}

	s, ok := c.protoCharacter.NeededSkills[name]
	if ok {
		s.Ct -= amt
		if s.Ct <= 0 {
			delete(c.protoCharacter.NeededSkills, name)
		}
	}
}

func (c *boltCharacter) IncrNeededItem(name string, amt uint64) {
	if c.protoCharacter.NeededItems == nil {
		c.protoCharacter.NeededItems = map[string]*ProtoItem{}
	}

	s, ok := c.protoCharacter.NeededItems[name]
	if !ok {
		c.protoCharacter.NeededItems[name] = &ProtoItem{Description: name, Count: amt}
	} else {
		s.Count += amt
	}
}

func (c *boltCharacter) DecrNeededItem(name string, amt uint64) {
	if c.protoCharacter.NeededItems == nil {
		return
	}

	s, ok := c.protoCharacter.NeededItems[name]
	if ok {
		s.Count -= amt
		if s.Count <= 0 {
			delete(c.protoCharacter.NeededItems, name)
		}
	}
}
