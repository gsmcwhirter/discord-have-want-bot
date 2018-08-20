package storage

type boltTransmute struct {
	protoTransm *ProtoTransmute
}

func (s boltTransmute) Name() string {
	return s.protoTransm.Name
}

func (s boltTransmute) Count() uint64 {
	return s.protoTransm.Count
}
