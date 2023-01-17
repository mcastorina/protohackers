package lrcp

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
)

type (
	lrcpMsg interface {
		SessionID() uint32
	}

	connectMsg struct {
		sessionID uint32
	}
	dataMsg struct {
		sessionID uint32
		pos       uint32
		data      string
	}
	ackMsg struct {
		sessionID uint32
		length    uint32
	}
	closeMsg struct {
		sessionID uint32
	}
)

func parseMsg(data []byte) (lrcpMsg, error) {
	invalidMsg := errors.New("invalid message format")
	if len(data) < 2 ||
		!bytes.HasPrefix(data, []byte("/")) ||
		!bytes.HasSuffix(data, []byte("/")) {
		return nil, invalidMsg
	}
	data = data[1 : len(data)-1]
	parts := strings.SplitN(string(data), "/", 4)

	if len(parts) < 2 {
		return nil, invalidMsg
	}
	var id uint32
	if num, err := strconv.Atoi(parts[1]); err == nil {
		id = uint32(num)
	} else {
		return nil, invalidMsg
	}

	switch parts[0] {
	case "connect":
		if len(parts) != 2 {
			return nil, invalidMsg
		}
		return connectMsg{id}, nil
	case "data":
		if len(parts) != 4 {
			return nil, invalidMsg
		}
		pos, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, invalidMsg
		}
		return dataMsg{
			sessionID: id,
			pos:       uint32(pos),
			data:      unescape(parts[3]),
		}, nil
	case "ack":
		if len(parts) != 3 {
			return nil, invalidMsg
		}
		length, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, invalidMsg
		}
		return ackMsg{sessionID: id, length: uint32(length)}, nil
	case "close":
		if len(parts) != 2 {
			return nil, invalidMsg
		}
		return closeMsg{id}, nil
	default:
		return nil, errors.New("unrecognized message")
	}
}

func (m connectMsg) SessionID() uint32 { return m.sessionID }
func (m dataMsg) SessionID() uint32    { return m.sessionID }
func (m ackMsg) SessionID() uint32     { return m.sessionID }
func (m closeMsg) SessionID() uint32   { return m.sessionID }

func unescape(s string) string {
	replacer := strings.NewReplacer("\\/", "/", "\\\\", "\\")
	return replacer.Replace(s)
}

func escape(s string) string {
	replacer := strings.NewReplacer("/", "\\/", "\\", "\\\\")
	return replacer.Replace(s)
}
