package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/valeriugold/vket/shared/jsonconfig"
	"github.com/valeriugold/vket/vfiles"
	"github.com/valeriugold/vket/vlog"
	model "github.com/valeriugold/vket/vmodel"
	"github.com/valeriugold/vket/vviews"
)

var store = sessions.NewCookieStore([]byte("secret-project"))

// config the settings variable
var config = &configuration{}

// configuration contains the application settings
type configuration struct {
	Database database.Info `json:"Database"`
	// Email     email.SMTPInfo  `json:"Email"`
	// Recaptcha recaptcha.Info  `json:"Recaptcha"`
	// Server   server.Server   `json:"Server"`
	Server Configuration        `json:"Server"`
	Log    vlog.Configuration   `json:"Log"`
	VFiles vfiles.Configuration `json:"VFiles"`
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
var serverConfig Configuration

// InitConfiguration: copy configuration to local config variable
func InitConfiguration(c Configuration) {
	serverConfig = c
	log.Printf("server: %v\n", serverConfig)
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

	vlog.InitConfiguration(config.Log)

	vlog.Trace.Println("Start logging trace")
	vlog.Warning.Printf("This is a warning")

	// configure http server
	InitConfiguration(config.Server)
	// Connect to database
	database.Connect(config.Database)
	// init vfiles
	vfiles.InitConfiguration(config.VFiles)

	vviews.Init()
	stdChain := alice.New(loggingSetter(os.Stdout))

	//static file handler
	// http.Handle("/bootstrap/", http.StripPrefix("/bootstrap/", http.FileServer(http.Dir("bootstrap-3.3.7-dist"))))
	http.Handle("/bootstrap/", http.StripPrefix("/bootstrap/", http.FileServer(http.Dir(serverConfig.Static))))
	http.Handle("/login", stdChain.ThenFunc(LoginALL))
	http.Handle("/register", stdChain.ThenFunc(RegisterALL))
	http.Handle("/about", stdChain.ThenFunc(AboutGET))
	http.Handle("/events", stdChain.ThenFunc(EventsGET))
	http.Handle("/newevent", stdChain.ThenFunc(NewEventALL))
	http.Handle("/hello", stdChain.ThenFunc(HelloGET))
	// http.Handle("/uploadmovies", stdChain.ThenFunc(UploadMoviesALL))
	http.Handle("/uploadforevent", stdChain.ThenFunc(UploadForEventALL))

	http.Handle("/logout", stdChain.ThenFunc(LogoutGET))
	http.Handle("/exitNow", stdChain.ThenFunc(ExitNowGET))

	err := http.ListenAndServe(":"+strconv.Itoa(serverConfig.Port), context.ClearHandler(http.DefaultServeMux))
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
	email := s.Values["email"].(string)
	user, err := model.UserByEmail(email)
	if err != nil {
		vlog.Warning.Printf("error on UserByEmail(%s) = %v", email, err)
	} else {
		vlog.Trace.Printf("user email %s = %v", email, user)
	}
	//FormatUint(, base int) string
	vviews.Hello(w, fmt.Sprintf("%d", s.Values["ID"].(uint32)),
		s.Values["email"].(string), user.Email, s.Values["role"].(string),
		user.Role, user.FirstName, user.LastName, user.Password,
		fmt.Sprintf("%d", user.ID))
}

func UploadForEventALL(w http.ResponseWriter, r *http.Request) {
	s, err := getAuthenticatedSession(w, r)
	if err != nil {
		return
	}
	s.Save(r, w)
	vlog.Trace.Printf("r.Method: %s\n", r.Method)

	if r.Method == "POST" {
		UploadForEventPOST(w, r)
	} else {
		UploadForEventGET(w, r)
	}
}

func UploadForEventGET(w http.ResponseWriter, r *http.Request) {
	s, err := getAuthenticatedSession(w, r)
	if err != nil {
		return
	}
	s.Save(r, w)
	eventID := r.FormValue("eventID")

	vviews.UploadMovies(w, eventID)
}

func UploadForEventPOST(w http.ResponseWriter, r *http.Request) {
	_, err := getAuthenticatedSession(w, r)
	if err != nil {
		return
	}
	// ID := s.Values["ID"].(uint32)

	vlog.Trace.Printf("Upload ...")
	//get the multipart reader for the request.
	reader, err := r.MultipartReader()
	if err != nil {
		vlog.Warning.Printf("MultpartReader, err=%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// eventID := r.FormValue("eventID")
	// vlog.Trace.Printf("converting ev=%v", eventID)
	// eid, err := strconv.ParseUint(eventID, 10, 32)
	// if err != nil {
	// 	vlog.Warning.Printf("eventID=%s is not integer, err=%v", eventID, err)
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	eid := 0
	vlog.Trace.Printf("calling SaveMultipart")
	if err = vfiles.SaveMultipart(uint32(eid), reader); err != nil {
		vlog.Warning.Printf("err on SaveMultipart, err:%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vlog.Trace.Printf("success!\n")
	//display success message.
	vviews.Hello(w, "upload", "successful")
}

// func UploadMoviesALL(w http.ResponseWriter, r *http.Request) {
// 	s, err := getAuthenticatedSession(w, r)
// 	if err != nil {
// 		return
// 	}
// 	s.Save(r, w)
// 	vlog.Trace.Printf("r.Method: %s\n", r.Method)
// 	if r.Method == "POST" {
// 		UploadMoviesPOST(w, r)
// 	} else {
// 		UploadMoviesGET(w, r)
// 	}
// }
// func UploadMoviesGET(w http.ResponseWriter, r *http.Request) {
// 	s, err := getAuthenticatedSession(w, r)
// 	if err != nil {
// 		return
// 	}
// 	s.Save(r, w)
// 	vviews.UploadMovies(w)
// }
// func UploadMoviesPOST(w http.ResponseWriter, r *http.Request) {
// 	s, err := getAuthenticatedSession(w, r)
// 	if err != nil {
// 		return
// 	}
// 	ID := s.Values["ID"].(uint32)
// 	//get the multipart reader for the request.
// 	reader, err := r.MultipartReader()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	if err = vfiles.SaveMultipart(ID, reader); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	vlog.Trace.Printf("success!\n")
// 	//display success message.
// 	vviews.Hello(w, "upload", "successful")
// }

func ExitNowGET(w http.ResponseWriter, r *http.Request) {
	os.Exit(0)
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
	s.Values["ID"] = user.ID
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

func EventsGET(w http.ResponseWriter, r *http.Request) {
	s, err := getAuthenticatedSession(w, r)
	if err != nil {
		return
	}
	s.Save(r, w)
	// get all events for this user
	userID := s.Values["ID"].(uint32)
	evs, err := model.EventGetAllForUserID(userID)
	if err != nil {
		vlog.Warning.Printf("Could not get events for user id %d, err:%v", userID, err)
		vviews.Error(w, "Could not get events for user id "+fmt.Sprintf("%d", userID)+" error="+err.Error())
		return
	}
	vviews.EventsShow(w, evs)
}
func NewEventALL(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		NewEventPOST(w, r)
	} else {
		NewEventGET(w, r)
	}
}
func NewEventGET(w http.ResponseWriter, r *http.Request) {
	vviews.NewEvent(w, "eventName")
}
func NewEventPOST(w http.ResponseWriter, r *http.Request) {
	s, err := getAuthenticatedSession(w, r)
	if err != nil {
		return
	}
	s.Save(r, w)

	evName := r.FormValue("eventName")
	userID := s.Values["ID"].(uint32)
	err = model.EventCreate(userID, evName)
	if err != nil {
		vlog.Warning.Printf("Create event for user id %d returned err=%s", userID, err.Error())
		// VG: show error page
		vviews.Error(w, "Create event for user id "+fmt.Sprintf("%d", userID)+" error="+err.Error())
		return
	}
	http.Redirect(w, r, "/events", http.StatusFound)
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
