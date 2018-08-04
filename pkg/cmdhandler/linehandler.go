package cmdhandler

// LineHandler TODOC
type LineHandler interface {
	HandleLine(user, guild string, line string) (string, error)
}

type lineHandlerFunc struct {
	handler func(user, guild string, line string) (string, error)
}

// NewLineHandler TODOC
func NewLineHandler(f func(string, string, string) (string, error)) LineHandler {
	return &lineHandlerFunc{handler: f}
}

func (lh *lineHandlerFunc) HandleLine(user, guild string, line string) (string, error) {
	return lh.handler(user, guild, line)
}
