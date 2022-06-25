package webui

import (
	"sync"

	"github.com/AckeeCZ/goproxie/internal/history"
	"github.com/AckeeCZ/goproxie/internal/util"
)

const (
	UIEventConnectionsChanged string = "connections-changed"
)

type UIStartedProxy struct {
	spawn *history.SpawnedHistoryCommand
	raw   string
	port  int
}

type UIState struct {
	startedProxies []*UIStartedProxy
	lock           sync.Mutex
	events         chan string
}

func (s *UIState) startHistoryCommandWithRaw(raw string) {
	localPort := history.ParseRaw(raw).LocalPort
	port := state.PortInfo(localPort)

	if !port.Available && !port.AvailableAfterProxyReplace {
		return
	}
	if port.AvailableAfterProxyReplace {
		s.stopHistoryCommandWithRaw(port.Proxy.raw)
	}
	go func() {
		spawn := history.ExecHistoryItem(raw)
		s.lock.Lock()
		s.startedProxies = append(s.startedProxies, &UIStartedProxy{
			spawn: spawn,
			raw:   raw,
			port:  localPort,
		})
		s.lock.Unlock()
		// Remove from history commands if it ends unexpectedly
		go func() {
			spawn.Wait()
			s.stopHistoryCommandWithRaw(raw)
			s.events <- UIEventConnectionsChanged
		}()
	}()
}

func (s *UIState) stopHistoryCommandWithRaw(raw string) {
	s.lock.Lock()
	for i, p := range s.startedProxies {
		if p.raw == raw {
			p.spawn.Kill()
			p.spawn.Wait()
			s.startedProxies[i] = s.startedProxies[len(s.startedProxies)-1]
			s.startedProxies = s.startedProxies[:len(s.startedProxies)-1]
		}
	}
	s.lock.Unlock()
}

var state = NewUIState()

type UIStatePortInfo struct {
	// Port is free to use
	Available bool
	// Port is used by a subprocess which can be ended easily
	AvailableAfterProxyReplace bool
	// Subprocess
	Proxy *UIStartedProxy
}

func (s *UIState) PortInfo(port int) UIStatePortInfo {
	available := util.IsPortFree(port)
	var proxy *UIStartedProxy = nil
	for _, p := range s.startedProxies {
		if p.port == port {
			proxy = p
		}
	}
	usedByUs := !available && proxy != nil
	return UIStatePortInfo{
		Available:                  available,
		AvailableAfterProxyReplace: usedByUs,
		Proxy:                      proxy,
	}
}

func (s *UIState) IsRawActive(raw string) bool {
	active := false
	for _, p := range s.startedProxies {
		if p.raw == raw {
			active = true
		}
	}
	return active
}

func NewUIState() *UIState {
	s := &UIState{}
	s.startedProxies = []*UIStartedProxy{}
	s.events = make(chan string, 10)
	return s
}
