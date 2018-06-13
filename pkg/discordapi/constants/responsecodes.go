package constants

// ResponseCode TODOC
type ResponseCode int

// ResponseCode names
const (
	UnknownError         ResponseCode = 4000
	UnknownOpcode                     = 4001
	DecodeError                       = 4002
	NotAuthenticated                  = 4003
	AuthenticationFailed              = 4004
	AlreadyAuthenticated              = 4005
	InvalidSequence                   = 4007
	RateLimited                       = 4008
	SessionTimeout                    = 4009
	InvalidShard                      = 4010
	ShardingRequired                  = 4011
)
