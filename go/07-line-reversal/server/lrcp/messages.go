package lrcp

type (
	lrcpMsg interface {
		SessionID() uint32
	}

	connectMsg struct {
	}
	dataMsg struct {
	}
	ackMsg struct {
	}
	closeMsg struct {
	}
)

// func (_ connectMsg) lrcpMsg() {}
// func (_ dataMsg) lrcpMsg()    {}
// func (_ ackMsg) lrcpMsg()     {}
// func (_ closeMsg) lrcpMsg()   {}

func parseMsg(data []byte) (lrcpMsg, error) {
	panic("todo")
}

func (_ connectMsg) SessionID() uint32 { panic("todo") }
func (_ dataMsg) SessionID() uint32    { panic("todo") }
func (_ ackMsg) SessionID() uint32     { panic("todo") }
func (_ closeMsg) SessionID() uint32   { panic("todo") }
