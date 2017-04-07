package vviews

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/valeriugold/vket/vlog"
)

// elements from bootstrap navbar, <href=Name>Display<>
type navItem struct {
	Name    string
	Display string
}

// for active element Active=Name, Navbar holds all the elements in navbar for current view
type nav struct {
	Active string
	Navbar []navItem
}

// load templates
// define generic json printing functions
// define how functions interract

// global var holding all templates
// var t *template.Template
var views = make(map[string]*View)
var nameOfBaseTmpl = "base"
var dir = "/Users/valeriug/dev/go/src/github.com/valeriugold/vket/vviews/vtemplates"

// Init should be called automatically when this package is used
func Init() {
	names := []string{"about", "error", "hello", "vuploadmovie",
		"login", "register", "newevent", "eventsshow", "filesshow", "fineuploader-s3-ui"}
	navActives := []string{"about", "error", "hello", "uploadmovie", "login", "register",
		"newevent", "eventsshow", "filesshow", "fineuploader-s3-ui"}
	navItems := []navItem{{"about", "About"},
		{"login", "Login"},
		{"register", "Register"},
		{"hello", "Hello"},
		// {"uploadmovies", "UploadMovies"},
		{"newevent", "NewEvent"},
		{"events", "Events"},
		{"logout", "Logout"},
		// {"fineuploader-s3-ui", "FineUploader"},
		{"exitNow", "Exit"}}
	for i, n := range names {
		views[n] = CreateView(n, nameOfBaseTmpl, []string{n}, navActives[i], navItems)
	}
}

func GetJSONRepresentation(args ...interface{}) string {
	// b, err := json.Marshal(args)
	b, err := json.MarshalIndent(args, "", "    ")
	if err != nil {
		vlog.Error.Println(err, " args=", args)
		return "error marshaling args"
	}
	return string(b)
}

// View defines a view, including the template files names and navbar
type View struct {
	name     string
	base     string
	files    []string
	navData  nav
	template *template.Template
}

func CreateView(name string, baseName string, files []string, navActive string, navItems []navItem) *View {
	fls := []string{baseName + ".tmpl"}
	for _, f := range files {
		fls = append(fls, f+".tmpl")
	}
	v := &View{name: name, base: baseName, files: fls, navData: nav{Active: navActive, Navbar: navItems}}
	v.Init()
	return v
}
func UseTemplate(w http.ResponseWriter, name string, data interface{}) {
	if v, ok := views[name]; ok {
		d := struct {
			Nav  nav
			Data interface{}
		}{v.navData, data}
		v.Render(w, d)
		return
	}
	// http.Error(w, "Did not find template name for data: %v", data)
	vlog.Error.Printf("Did not find template name !%s! for data: %v\n", name, data)
}
func (v *View) Init() {
	paths := make([]string, 0, len(v.files))
	for _, f := range v.files {
		vlog.Trace.Printf("d=%s, f=%v\n", dir, f)
		paths = append(paths, filepath.Join(dir, f))
	}
	vlog.Trace.Println("l=", len(paths), " paths = ", paths)
	vlog.Trace.Printf("0=%s!\n", paths[0])
	v.template = template.Must(template.ParseFiles(paths...))
}
func (v *View) Render(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := v.template.ExecuteTemplate(w, v.base, data)
	if err != nil {
		vlog.Error.Printf("err executing template %s: %v", v.name, err)
		return
	}
}

// type VView struct {
// 	name string
// }
//
// func GetVView() VView {
// 	return VView{}
// }

func About(w http.ResponseWriter) {
	UseTemplate(w, "about", nil)
}

func Hello(w http.ResponseWriter, fields ...string) {
	UseTemplate(w, "hello", fields)
}

func UploadMovies(w http.ResponseWriter, event interface{}) {
	UseTemplate(w, "vuploadmovie", event)
}
func Login(w http.ResponseWriter, fields ...string) {
	UseTemplate(w, "login", fields)
}

func Error(w http.ResponseWriter, fields ...string) {
	UseTemplate(w, "error", fields)
}

func Register(w http.ResponseWriter, fields ...string) {
	UseTemplate(w, "register", fields)
}

func NewEvent(w http.ResponseWriter, fields ...string) {
	UseTemplate(w, "newevent", fields)
}

func EventsShow(w http.ResponseWriter, events interface{}) {
	UseTemplate(w, "eventsshow", events)
}
func FielesShow(w http.ResponseWriter, event interface{}, files interface{}) {
	UseTemplate(w, "filesshow", struct {
		Event interface{}
		Files interface{}
	}{Event: event, Files: files})
}
func FineUploadMovies(w http.ResponseWriter, event interface{}) {
	vlog.Trace.Printf("show fineuploader-s3-ui, event=%v", event)
	UseTemplate(w, "fineuploader-s3-ui", event)
}
