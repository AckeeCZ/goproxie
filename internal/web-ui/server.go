package webui

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AckeeCZ/goproxie/internal/fsconfig"
	"golang.org/x/net/websocket"
)

var server *http.Server
var serverLogger = log.New(os.Stdout, "webserver: ", log.LstdFlags)

// Return HTTP handler loading a JavaScript file relative to internal/web-ui/asset
func staticJavaScriptController(path string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/javascript")
		staticController(filepath.Join("asset", path))(res, req)
	}
}

func staticFaviconController() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		staticController(filepath.Join("asset", "favicon.ico"))(res, req)
	}
}

func staticController(path string) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		file, err := os.ReadFile(filepath.Join("internal", "web-ui", path))
		if err != nil {
			serverLogger.Printf("Failed to load static file %s: %s", path, err)
			res.WriteHeader(404)
			return
		}
		res.Write(file)
	}
}

// Starts UI HTTP server
func Start() {
	fsconfig.Initialize()
	handler := http.DefaultServeMux
	handler.HandleFunc("/javascript.js", staticJavaScriptController("javascript.js"))
	handler.HandleFunc("/server.js", staticJavaScriptController("server.js"))
	handler.HandleFunc("/favicon.ico", staticFaviconController())
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
	go func() {
		for event := range state.events {
			realtime.Write(event, event)
		}
	}()
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
