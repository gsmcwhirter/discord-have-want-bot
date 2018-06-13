package constants

// OpCode TODOC
type OpCode int

// OpCode names
const (
	Heartbeat      OpCode = 1
	HeartbeatAck          = 11
	Identify              = 2
	InvalidSession        = 9
	Resume                = 6
	Dispatch              = 0
	StatusUpdate          = 3
	Reconnect             = 7
	Hello                 = 10
)
