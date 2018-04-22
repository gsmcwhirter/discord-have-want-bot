package storage

type boltSkill struct {
	protoSkill *ProtoSkill
}

func (s boltSkill) Name() string {
	return s.protoSkill.Name
}

func (s boltSkill) Points() uint64 {
	return s.protoSkill.Ct
}
