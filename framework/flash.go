package framework

import (
	"encoding/gob"

	"github.com/gorilla/sessions"
)

type FlashType uint8

const (
	FlashSuccess FlashType = iota
	FlashInfo
	FlashWarn
	FlashFail
)

/*
 * Flash represents a cross-page message to the user. It typically appears once
 * and then is erased.
 */
type Flash struct {
	Type    FlashType
	Message string
}

func Flashes(session *sessions.Session) []Flash {

	rawFlashes := session.Flashes()

	if len(rawFlashes) == 0 {
		return nil
	}

	var flashes []Flash

	for _, flash := range rawFlashes {
		flashes = append(flashes, flash.(Flash))
	}

	return flashes
}

func init() {
	// register Flash so that we can use it with session
	gob.Register(Flash{})
}
