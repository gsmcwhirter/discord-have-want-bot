package cmdhandler

// LineHandler TODOC
type LineHandler interface {
	HandleLine(user string, line []rune) (string, error)
}

type lineHandlerFunc struct {
	handler func(user string, line []rune) (string, error)
}

// NewLineHandler TODOC
func NewLineHandler(f func(string, []rune) (string, error)) LineHandler {
	return lineHandlerFunc{handler: f}
}

func (lh lineHandlerFunc) HandleLine(user string, line []rune) (string, error) {
	return lh.handler(user, line)
}
