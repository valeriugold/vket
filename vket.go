package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/sessions"
	"github.com/justinas/alice"
	"github.com/valeriugold/vket/shared/database"
	"github.com/valeriugold/vket/vlog"
	model "github.com/valeriugold/vket/vmodel"
	"github.com/valeriugold/vket/vviews"
)

var store = sessions.NewCookieStore([]byte("secret-project"))
var config configuration

// config the settings variable
var config = &configuration{}

// configuration contains the application settings
type configuration struct {
	Database database.Info `json:"Database"`
	// Email     email.SMTPInfo  `json:"Email"`
	// Recaptcha recaptcha.Info  `json:"Recaptcha"`
	// Server   server.Server   `json:"Server"`
	Server Configuration      `json:"Server"`
	Log    vlog.Configuration `json:"Log"`
	// Session  session.Session `json:"Session"`
	// Template view.Template   `json:"Template"`
	// View     view.View       `json:"View"`
}

// ParseJSON unmarshals bytes to structs
func (c *configuration) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
}

type Configuration struct {
	Port   int    `json:"port"`
	Static string `json:"static"`
}

// configuration values for the server
var server = &config.Server

// InitConfiguration: parse config section and apply the configuration
func InitConfiguration(jb []byte) {
	var s section
	s.Section = &config
	// default values
	config.Port = 9090
	err := json.Unmarshal(jb, &s)
	if err != nil {
		log.Fatal("Config Parse Error:", err)
	}
	log.Printf("server: %v\n", config)
}

func loggingSetter(out io.Writer) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(out, h)
	}
}

// load JSON conf file and filter out comments
func loadConfiguration(path string) []byte {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Config File ", path, " is missing", err)
	}
	var reComments = regexp.MustCompile(`(?m)^\s*//.*$\n?`)
	b := reComments.ReplaceAll(f, []byte{})
	log.Printf("no comments %v", string(b))
	return b
}

func main() {
	var configFile = "./vket.json"
	// var configFile = "config"+string(os.PathSeparator)+"config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	// Load the configuration file
	jsonconfig.Load(configFile, config)

	confBytes := loadConfiguration(configFile)
	config.ParseJSON(b []byte)
	
	InitConfiguration(confBytes)
	vlog.InitConfiguration(confBytes)
	// os.Exit(0)
	// vlog.Init(os.Stdout, os.Stdout, os.Stdout, os.Stderr)

	vlog.Trace.Println("Start logging trace")
	vlog.Warning.Printf("This is a warning")

	// Connect to database
	database.Connect(config.Database)

	vviews.Init()
	stdChain := alice.New(loggingSetter(os.Stdout))

	//static file handler
	// http.Handle("/bootstrap/", http.StripPrefix("/bootstrap/", http.FileServer(http.Dir("bootstrap-3.3.7-dist"))))
	http.Handle("/bootstrap/", http.StripPrefix("/bootstrap/", http.FileServer(http.Dir(config.Static))))
	http.Handle("/login", stdChain.ThenFunc(LoginALL))
	http.Handle("/register", stdChain.ThenFunc(RegisterALL))
	http.Handle("/about", stdChain.ThenFunc(AboutGET))
	http.Handle("/hello", stdChain.ThenFunc(HelloGET))
	http.Handle("/uploadmovies", stdChain.ThenFunc(UploadMoviesALL))
	http.Handle("/logout", stdChain.ThenFunc(LogoutGET))
	err := http.ListenAndServe(":"+strconv.Itoa(config.Port), context.ClearHandler(http.DefaultServeMux))
	// err := http.ListenAndServe(":9090", context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		vlog.Error.Fatal("ListenAndServe: ", err)
	}
}

func getAuthenticatedSession(w http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	s, err := store.Get(r, "session-x")
	if err != nil {
		vlog.Warning.Println("err on getting session ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	if s.Values["authenticated"] == nil || s.Values["authenticated"].(string) != "yes" {
		vlog.Info.Println("not authenticated")
		vviews.Error(w, "the context is not authenticated")
		return s, errors.New("not authenticated")
	}
	return s, nil
}

// AboutGET controller section
func AboutGET(w http.ResponseWriter, r *http.Request) {
	vviews.About(w)
}

func HelloGET(w http.ResponseWriter, r *http.Request) {
	s, err := getAuthenticatedSession(w, r)
	if err != nil {
		return
	}
	s.Save(r, w)
	vviews.Hello(w, s.Values["email"].(string), s.Values["role"].(string))
	email := s.Values["email"].(string)
	user, err := model.UserByEmail(email)
	if err != nil {
		vlog.Warning.Printf("error on UserByEmail(%s) = %v", email, err)
	} else {
		vlog.Trace.Printf("user email %s = %v", email, user)
	}
}

func UploadMoviesALL(w http.ResponseWriter, r *http.Request) {
	s, err := getAuthenticatedSession(w, r)
	if err != nil {
		return
	}
	s.Save(r, w)
	vlog.Trace.Printf("r.Method: %s\n", r.Method)

	if r.Method == "POST" {
		UploadMoviesPOST(w, r)
	} else {
		UploadMoviesGET(w, r)
	}
}

func UploadMoviesGET(w http.ResponseWriter, r *http.Request) {
	s, err := getAuthenticatedSession(w, r)
	if err != nil {
		return
	}
	s.Save(r, w)
	vviews.UploadMovies(w)
}

func UploadMoviesPOST(w http.ResponseWriter, r *http.Request) {
	//get the multipart reader for the request.
	reader, err := r.MultipartReader()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//copy each part to destination.
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		//if part.FileName() is empty, skip this iteration.
		if part.FileName() == "" {
			continue
		}
		vlog.Trace.Printf("file: %s\n", part.FileName())
		dst, err := os.Create("/tmp/" + part.FileName())
		defer dst.Close()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(dst, part); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	vlog.Trace.Printf("success!\n")
	//display success message.
	vviews.Hello(w, "upload", "successful")
}

func LogoutGET(w http.ResponseWriter, r *http.Request) {
	s, err := getAuthenticatedSession(w, r)
	if err != nil {
		return
	}
	// Clear out all stored values in the cookie
	for k := range s.Values {
		delete(s.Values, k)
	}
	s.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func LoginALL(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		LoginPOST(w, r)
	} else {
		LoginGET(w, r)
	}
}
func LoginGET(w http.ResponseWriter, r *http.Request) {
	vviews.Login(w, "email", "password")
}

func LoginPOST(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := model.UserByEmail(email)
	if err != nil {
		vlog.Warning.Printf("error getting user with email %s, err=%v", email, err)
		vviews.Error(w, "error getting user with email "+email+", err="+err.Error())
		return
	}
	if user.Password != password {
		vlog.Warning.Printf("wrong pass for user with email %s", email)
		vviews.Error(w, "wrong pass for user with email "+email)
		return
	}
	s, err := store.Get(r, "session-x")
	if err != nil {
		vlog.Warning.Println("err on getting session ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.Values["email"] = email
	s.Values["authenticated"] = "yes"
	s.Values["role"] = user.Role
	s.Values["firstName"] = user.FirstName
	s.Values["lastName"] = user.LastName
	s.Save(r, w)
	http.Redirect(w, r, "/hello", http.StatusFound)
}

func RegisterALL(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		RegisterPOST(w, r)
	} else {
		RegisterGET(w, r)
	}
}
func RegisterGET(w http.ResponseWriter, r *http.Request) {
	vviews.Register(w, "firstName", "lastName", "email", "password", "role")
}

func RegisterPOST(w http.ResponseWriter, r *http.Request) {
	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")
	email := r.FormValue("email")
	password := r.FormValue("password")
	role := r.FormValue("role")
	err := model.UserCreate(firstName, lastName, email, password, role)
	if err != nil {
		vlog.Warning.Printf("Create user %s returned err=%s", email, err.Error())
		// VG: show error page
		vviews.Error(w, "error or user already exits for email="+email+" error="+err.Error())
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

// func (w http.ResponseWriter, r *http.Request) {
// }

type Person struct {
	user     string
	email    string
	password string
}

// type Models map[string]Person

// var model = Models{}

// var (
// 	ModelErrWrongPassword = errors.New("Wrong password.")
// 	ModelErrUserNotFound  = errors.New("User not found.")
// 	ModelErrUnkown        = errors.New("Generic misterious error.")
// )

// func (m *Models) SaveNewUser(email string, password string, user string) bool {
// 	if _, ok := (*m)[email]; ok {
// 		return false
// 	}
// 	(*m)[email] = Person{user, email, password}
// 	return true
// }

// func (m *Models) ChecUserPassword(email string, password string) (string, error) {
// 	if p, ok := (*m)[email]; ok {
// 		if p.password == password {
// 			return p.user, nil
// 		}
// 		return "", ModelErrWrongPassword
// 	}
// 	return "", ModelErrUserNotFound
// }
