package webui

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/AckeeCZ/goproxie/internal/fsconfig"
	"golang.org/x/net/websocket"
)

var server *http.Server
var serverLogger = log.New(os.Stdout, "webserver: ", log.LstdFlags)

// Starts UI HTTP server
func Start() {
	fsconfig.Initialize()
	handler := http.DefaultServeMux
	// TODO Refactor assets to sth like assets({ javascript: [...], css: [...] })
	handler.HandleFunc("/javascript.js", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/javascript")
		page.JavaScriptFile("./javascript.js", rw)
	})
	handler.HandleFunc("/server.js", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/javascript")
		page.JavaScriptFile("./server.js", rw)
	})
	handler.HandleFunc("/favicon.ico", func(rw http.ResponseWriter, r *http.Request) {
		page.Favicon(rw)
	})
	handler.HandleFunc("/history-commands-list", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(page.fragment.HistoryCommandList(r.URL.Query().Get("query"))))
	})
	handler.HandleFunc("/connect-history-item", func(rw http.ResponseWriter, r *http.Request) {
		// 💡 Struct's props has to be with first-capital if you want json.unmarshall to be
		// able to access it and set it
		requestBody := struct {
			Raw   string `json:"raw"`
			Query string `json:"query"`
		}{}
		if err := jsonUnmarshallBody(r, &requestBody); err != nil {
			log.Fatal(err)
		}
		state.startHistoryCommandWithRaw(requestBody.Raw)
		rw.Write([]byte(page.fragment.HistoryCommandList(requestBody.Query)))
	})
	handler.HandleFunc("/disconnect-history-item", func(rw http.ResponseWriter, r *http.Request) {
		requestBody := struct {
			Raw   string `json:"raw"`
			Query string `json:"query"`
		}{}
		if err := jsonUnmarshallBody(r, &requestBody); err != nil {
			log.Fatal(err)
		}
		state.stopHistoryCommandWithRaw(requestBody.Raw)
		rw.Write([]byte(page.fragment.HistoryCommandList(requestBody.Query)))
	})
	handler.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		page.Main(rw, struct{ searchQuery string }{r.URL.Query().Get("query")})
	})
	realtime := NewRealtime()
	realtime.onMessage = func(c *websocket.Conn, m *RealtimeMessage) *RealtimeMessage {
		if m.Type == realtimeMessageType.historySearch {
			return &RealtimeMessage{
				Type: realtimeMessageType.refresh,
			}
		}
		return nil
	}
	handler.Handle("/rt", realtime.WebSocketHandler(func() {

	}))
	server = &http.Server{
		Handler: logged(handler),
	}
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Goproxie Web-UI started. Visit http://%s\n", listener.Addr().String())
	server.Serve(listener)
}

func jsonUnmarshallBody(r *http.Request, out interface{}) error {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &out)
	if err != nil {
		return err
	}
	return nil
}

func logged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		serverLogger.Printf("%s %s", req.Method, req.URL)
		next.ServeHTTP(res, req)
	})
}
