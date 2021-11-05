package webui

import (
	"context"
	"sync"

	"github.com/AckeeCZ/goproxie/internal/history"
)

type UIState struct {
	historyRawToIsActive map[string]bool
	rawToCommandCancel   map[string]context.CancelFunc
	lock                 sync.Mutex
}

func (s *UIState) startHistoryCommandWithRaw(raw string) {
	s.historyRawToIsActive[raw] = true
	go func() {
		c := history.ExecHistoryItem(raw)
		s.lock.Lock()
		s.rawToCommandCancel[raw] = c
		s.lock.Unlock()
	}()

}

func (s *UIState) stopHistoryCommandWithRaw(raw string) {
	s.historyRawToIsActive[raw] = false
	c := s.rawToCommandCancel[raw]
	if c == nil {
		return
	}
	c()
}

var state = NewUIState()

func NewUIState() *UIState {
	s := &UIState{}
	s.historyRawToIsActive = map[string]bool{}
	s.rawToCommandCancel = map[string]context.CancelFunc{}
	return s
}
