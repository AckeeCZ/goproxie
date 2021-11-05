package webui

import (
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/AckeeCZ/goproxie/internal/history"
	"github.com/AckeeCZ/goproxie/internal/util"
	"github.com/AckeeCZ/goproxie/internal/version"
)

type WebUIPage struct {
	fragment HTMLFragment
}

var page = WebUIPage{}

var tmpl *template.Template

func (p *WebUIPage) LoadTemplates() {
	templates := []string{
		"internal/web-ui/main.html",
		"internal/web-ui/history-commands-list.html",
		"internal/web-ui/history-command-search.html",
	}
	t, err := template.ParseFiles(templates...)
	if err != nil {
		log.Fatal(err)
	}
	tmpl = t
}

type myWriter struct {
	bytes []byte
}

func (w *myWriter) Write(p []byte) (n int, err error) {
	w.bytes = append(w.bytes, p...)
	return len(p), nil
}
func (w *myWriter) toHTML() template.HTML {
	return template.HTML(string(w.bytes))
}

func executeTemplate(wr io.Writer, name string, data interface{}) {
	page.LoadTemplates()
	tmpl.ExecuteTemplate(wr, name, data)
}

func executeTemplateAsHTML(name string, data interface{}) template.HTML {
	page.LoadTemplates()
	wr := myWriter{}
	tmpl.ExecuteTemplate(&wr, name, data)
	return wr.toHTML()
}

type HTMLFragment struct {
}

func (*HTMLFragment) HistoryCommandSearch(searchQuery string) template.HTML {
	data := &map[string]interface{}{
		"query": searchQuery,
	}
	return executeTemplateAsHTML("history-command-search.html", data)
}

func (*HTMLFragment) HistoryCommandList(query string) template.HTML {
	list := history.List()
	raws := make([]string, len(list))
	for i, r := range list {
		raws[i] = r.Raw
	}
	filteredRaws := util.FilterStrings(raws, query)
	// 💡 Struct's props has to be with first-capital if you want template.HTML to be
	// able to access it and read it
	type it struct {
		Command history.Item
		Active  bool
	}
	items := make([]it, len(filteredRaws))
	for i, raw := range filteredRaws {
		for _, item := range list {
			if item.Raw == raw {
				items[i] = it{
					Command: item,
					Active:  state.historyRawToIsActive[item.Raw],
				}
				break
			}
		}
	}
	return executeTemplateAsHTML(
		"history-commands-list.html",
		&map[string]interface{}{
			"items": items,
			"query": query,
		},
	)
}

func (p *WebUIPage) Main(wr io.Writer, param struct {
	searchQuery string
}) {
	var data interface{} = map[string]interface{}{
		"version":              version.Get(),
		"historyCommands":      p.fragment.HistoryCommandList(param.searchQuery),
		"historyCommandSearch": p.fragment.HistoryCommandSearch(param.searchQuery),
	}
	executeTemplate(wr, "main.html", data)
}

func (p *WebUIPage) JavaScriptFile(path string, wr io.Writer) {
	p.Asset(path, wr)
}

func (p *WebUIPage) Asset(path string, wr io.Writer) {
	res, err := os.ReadFile(filepath.Join("internal", "web-ui", "asset", path))
	if err != nil {
		serverLogger.Printf("Failed to load asset file %s: %s", path, err)
		return
	}
	wr.Write(res)
}

func (p *WebUIPage) Favicon(wr io.Writer) {
	p.Asset("favicon.ico", wr)
}
