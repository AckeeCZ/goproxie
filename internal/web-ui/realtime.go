package webui

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

var mutex = &sync.Mutex{}
var realtimeLogger = log.New(io.Discard, "webserver/realtime: ", log.LstdFlags)

type websocketKeepAliveReader struct {
	// Dummy channel to wait forever in Read to block the Reader and keep WS alive
	wait chan int
	// When written to quit, Read returns 0, closing the Reader and the WS conn
	quit chan int
}

func (r *websocketKeepAliveReader) Close() {
	realtimeLogger.Println("Closing socket")
	r.quit <- 0
}

func (r *websocketKeepAliveReader) Read(p []byte) (i int, err error) {
	realtimeLogger.Println("New socket opened")
	for {
		select {
		case <-r.wait:
		case <-r.quit:
			realtimeLogger.Println("Socket closed")
			return 0, io.EOF
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

type RealtimeMessage struct {
	V    int         `json:"v"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

var realtimeMessageType = struct {
	historySearch string
	refresh       string
}{historySearch: "history-search", refresh: "refresh"}

type Realtime struct {
	// Array of WS connections
	websocketConnections []*websocket.Conn
	// WS dummy ready for every WS conn to keep it open. See WebSocketHandler
	websocketConnToReader map[*websocket.Conn]*websocketKeepAliveReader
	onMessage             func(*websocket.Conn, *RealtimeMessage) *RealtimeMessage
}

func (r *Realtime) Write(messageType string, data interface{}) {
	message := &RealtimeMessage{V: 1, Type: messageType, Data: data}
	bytes, err := json.Marshal(message)
	if err != nil {
		realtimeLogger.Println("Failed to serialize message", err)
		return
	}
	if len(r.websocketConnections) > 0 {
		realtimeLogger.Println("Sending message")
	}
	mutex.Lock()
	for i, ws := range r.websocketConnections {
		if ws == nil {
			continue
		}
		_, err := ws.Write(bytes)

		if err != nil {
			realtimeLogger.Println("Failed to write to WS connection, reason: ", err)
			// Clear connection
			r.websocketConnToReader[ws].Close()
			r.websocketConnections[i] = nil
			r.websocketConnToReader[ws] = nil
			// At this point, r.websoketConnections will may contain nil
			// connections, e.g. [nil, nil, <conn>]. These are removed
			// manually in WS handler, after writer is closed.
		}
	}
	mutex.Unlock()
}

func (r *Realtime) WebSocketHandler(onNewConnection func()) websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		go func() {
			buff := make([]byte, 255)
			for {
				l, err := ws.Read(buff[0:])
				if err != nil {
					break
				}
				realtimeLogger.Printf("Incoming message %s", string(buff[0:l]))
				m := RealtimeMessage{}
				json.Unmarshal(buff[0:l], &m)
				replyMessage := r.onMessage(ws, &m)
				if replyMessage != nil {
					r.Write(replyMessage.Type, &replyMessage.Data)
				}
			}
		}()
		mutex.Lock()
		r.websocketConnections = append(r.websocketConnections, ws)

		realtimeLogger.Println(fmt.Sprintf("Incoming websocket connection (%d connections)", len(r.websocketConnections)))
		reader := &websocketKeepAliveReader{quit: make(chan int)}
		r.websocketConnToReader[ws] = reader
		mutex.Unlock()
		go onNewConnection()
		// Without next line, copying WS connection to a never-ending reader,
		// WS connection gets closed automatically
		// This is a blocking call; until reader gets closed
		io.Copy(ws, reader)
		// Clear dead connections from connection pool
		mutex.Lock()
		realtimeLogger.Println(fmt.Sprintf("Closed websocket connection (%d connections)", len(r.websocketConnections)-1))
		var conns []*websocket.Conn
		for _, conn := range r.websocketConnections {
			if conn == nil {
				continue
			}
			conns = append(conns, conn)
		}
		r.websocketConnections = conns
		mutex.Unlock()
	})
}

func NewRealtime() *Realtime {
	if strings.Contains(os.Getenv("DEBUG"), "realtime") {
		realtimeLogger = log.New(os.Stdout, "webserver/realtime: ", log.LstdFlags)
	}
	rt := Realtime{}
	rt.websocketConnToReader = make(map[*websocket.Conn]*websocketKeepAliveReader)
	rt.websocketConnections = make([]*websocket.Conn, 0)
	// Time ticker, send time to realtime clients every X seconds
	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond)
		for x := range ticker.C {
			rt.Write("time", fmt.Sprintf("%d", x.Unix()))
		}
	}()
	return &rt
}
