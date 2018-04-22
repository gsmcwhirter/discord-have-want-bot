package storage

type boltItem struct {
	protoItem *ProtoItem
}

func (s boltItem) Name() string {
	return s.protoItem.Description
}

func (s boltItem) Count() uint64 {
	return s.protoItem.Count
}
