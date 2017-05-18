package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/sessions"
	"github.com/justinas/alice"
	"github.com/valeriugold/vket/app/shared/database"
	"github.com/valeriugold/vket/app/shared/jsonconfig"
	"github.com/valeriugold/vket/app/shared/vlog"
	"github.com/valeriugold/vket/app/vcloud/vs3"
	"github.com/valeriugold/vket/app/vcloud/vs3/vfineuploader"
	"github.com/valeriugold/vket/app/vmodel"
	"github.com/valeriugold/vket/app/vviews"
)

var uploader = vfineuploader.New()
var vs = vs3.New()
var vr = vmodel.New(vs)

var store = sessions.NewCookieStore([]byte("secret-project"))

// VSession holds all user data from a user session
type VSession struct {
	Authenticated bool
	UserID        uint32
	Email         string
	Role          string
	FirstName     string
	LastName      string
}

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
	// VFiles vfiles.Configuration `json:"VFiles"`
	// Session  session.Session `json:"Session"`
	// Template view.Template   `json:"Template"`
	// View     view.View       `json:"View"`
}

// ParseJSON unmarshals bytes to structs
func (c *configuration) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
}

// Configuration holds configuration for vket http service
type Configuration struct {
	Port         int    `json:"Port"`
	Static       string `json:"Static"`
	FineUploader string `json:"FineUploader"`
}

// configuration values for the server
var serverConfig Configuration

// InitConfiguration copy configuration to local config variable
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

	// uploader := vfineuploader.New()
	// vr := vmodel.New()
	// vs := vs3.New()

	vviews.Init()
	stdChain := alice.New(loggingSetter(os.Stdout))

	//static file handler
	// http.Handle("/bootstrap/", http.StripPrefix("/bootstrap/", http.FileServer(http.Dir("bootstrap-3.3.7-dist"))))
	http.Handle("/bootstrap/", http.StripPrefix("/bootstrap/", http.FileServer(http.Dir(serverConfig.Static))))
	http.Handle("/s3.fine-uploader/", http.StripPrefix("/s3.fine-uploader/", http.FileServer(http.Dir(serverConfig.FineUploader))))
	http.Handle("/", stdChain.ThenFunc(AboutGET))
	http.Handle("/login", stdChain.ThenFunc(LoginALL))
	http.Handle("/register", stdChain.ThenFunc(RegisterALL))
	http.Handle("/about", stdChain.ThenFunc(AboutGET))
	http.Handle("/events", stdChain.ThenFunc(EventsGET))
	http.Handle("/newevent", stdChain.ThenFunc(NewEventALL))
	http.Handle("/hello", stdChain.ThenFunc(HelloGET))
	// http.Handle("/uploadmovies", stdChain.ThenFunc(UploadMoviesALL))
	http.Handle("/fineuploader-s3-ui", stdChain.ThenFunc(FineUploadForEventGET))
	http.Handle("/filesop", stdChain.ThenFunc(FilesOpPOST))

	http.Handle("/editorevents", stdChain.ThenFunc(EditorEventsGET))

	http.Handle("/files", stdChain.ThenFunc(FilesGET))
	http.Handle("/logout", stdChain.ThenFunc(LogoutGET))
	http.Handle("/exitNow", stdChain.ThenFunc(ExitNowGET))
	http.Handle("/upldsign", stdChain.ThenFunc(UploadSignPOST))
	http.Handle("/upldresultsuccess", stdChain.ThenFunc(UploadResultSuccess))

	err := http.ListenAndServe(":"+strconv.Itoa(serverConfig.Port), context.ClearHandler(http.DefaultServeMux))
	// err := http.ListenAndServe(":9090", context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		vlog.Error.Fatal("ListenAndServe: ", err)
	}
}

func authenticatedSessionSet(w http.ResponseWriter, r *http.Request, vsess VSession) (*sessions.Session, error) {
	s, err := store.Get(r, "session-x")
	if err != nil {
		vlog.Warning.Println("err on getting session ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return s, err
	}

	s.Values["ID"] = vsess.UserID
	s.Values["authenticated"] = vsess.Authenticated
	s.Values["email"] = vsess.Email
	s.Values["role"] = vsess.Role
	s.Values["firstName"] = vsess.FirstName
	s.Values["lastName"] = vsess.LastName
	s.Save(r, w)
	return s, nil
}

func authenticatedSessionClear(w http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	s, _, err := authenticatedSessionGet(w, r)
	if err != nil {
		return s, err
	}
	for k := range s.Values {
		delete(s.Values, k)
	}
	s.Save(r, w)
	return s, nil
}

func authenticatedSessionGet(w http.ResponseWriter, r *http.Request) (*sessions.Session, *VSession, error) {
	s, err := store.Get(r, "session-x")
	if err != nil {
		vlog.Warning.Println("err on getting session ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, nil, err
	}

	_, ok := s.Values["authenticated"]
	if !ok || !s.Values["authenticated"].(bool) {
		vlog.Info.Println("not authenticated")
		vviews.Error(w, "the context is not authenticated")
		return s, nil, errors.New("not authenticated")
	}
	// get all values for this session
	var vsess VSession
	vsess.Authenticated = true
	vsess.UserID = s.Values["ID"].(uint32)
	vsess.Email = s.Values["email"].(string)
	vsess.Role = s.Values["role"].(string)
	vsess.FirstName = s.Values["firstName"].(string)
	vsess.LastName = s.Values["lastName"].(string)
	s.Save(r, w)
	return s, &vsess, nil
}

// AboutGET controller section
func AboutGET(w http.ResponseWriter, r *http.Request) {
	vviews.About(w)
}

func FilesOpPOST(w http.ResponseWriter, r *http.Request) {
	_, vsess, err := authenticatedSessionGet(w, r)
	if err != nil {
		return
	}

	// check if user can acccess the files
	ev, err := GetEventFromFormCheckAccess(r, vsess)
	if err != nil {
		vlog.Warning.Printf("Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	areEditedFiles := false
	// form's filestype can be original or edited
	if filesType := r.FormValue("filesType"); filesType == "edited" {
		areEditedFiles = true
	}

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))

	r.ParseForm()
	vlog.Info.Printf("PostForm = %v\n", r.PostForm)
	if files, ok := r.PostForm["file"]; ok {
		vlog.Info.Printf("these are files' values: %v", files)
		efids := make([]uint32, len(files))
		for i, f := range files {
			efids[i], err = stringToUint32(f)
			if err != nil {
				vlog.Warning.Printf("event file ID %s is not integer, err=%v", f, err)
				return
			}
		}
		if a, ok := r.PostForm["action"]; ok {
			vlog.Info.Printf("action = %v", a)
			if len(a) == 1 {
				// get the slice on which the operation is allowed
				if areEditedFiles {
					efids = vmodel.GetEditedFidsAllowedOp(a[0], vsess.UserID, vsess.Role, efids)
				} else {
					efids = vmodel.GetOriginalFidsAllowedOp(a[0], vsess.UserID, vsess.Role, efids)
				}
				if a[0] == "delete" {
					vlog.Info.Printf("Delete files id %v", efids)
					for _, efid := range efids {
						if areEditedFiles {
							vlog.Info.Printf("delete edited file ID =%d", efid)
							if err = vr.DeleteDataByEditedFileID(efid); err != nil {
								vlog.Warning.Printf("could not delete edited file ID =%d, err=%v", efid, err)
								http.Error(w, err.Error(), http.StatusInternalServerError)
								return
							}
						} else {
							vlog.Info.Printf("delete event file ID =%d", efid)
							if err = vr.DeleteDataByEventFileID(efid); err != nil {
								vlog.Warning.Printf("could not delete event file ID =%d, err=%v", efid, err)
								http.Error(w, err.Error(), http.StatusInternalServerError)
								return
							}
						}
					}
					// show the files page without the deleted files
					efs, err := vmodel.EventFileGetAllForEventID(ev.ID)
					if err != nil {
						vlog.Warning.Printf("Could not get files for events id %d, err:%v", ev.ID, err)
						vviews.Error(w, "Could not get events for user id "+fmt.Sprintf("%d", ev.ID)+" error="+err.Error())
						return
					}

					vlog.Info.Printf("FielesShow: ev=%v, efs=%v", ev, efs)
					// http.Redirect(w, r, "/files?eventID=15", 303)
					RunDisplayFiles(w, ev, vsess)
					// vviews.FielesShow(w, ev, efs, true)
				} else if a[0] == "download" {
					vlog.Info.Printf("Download files id %v", files)
					zpr := vs.GetZipper()
					if err = vr.DownloadFiles(w, r, ev.ID, areEditedFiles, efids, zpr); err != nil {
						vlog.Warning.Printf("could not download, err=%v", err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}
			}
		}
	}
}

func HelloGET(w http.ResponseWriter, r *http.Request) {
	_, vsess, err := authenticatedSessionGet(w, r)
	if err != nil {
		return
	}
	user, err := vmodel.UserByEmail(vsess.Email)
	if err != nil {
		vlog.Warning.Printf("error on UserByEmail(%s) = %v", vsess.Email, err)
	} else {
		vlog.Trace.Printf("user email %s = %v", vsess.Email, user)
	}
	//FormatUint(, base int) string
	vviews.Hello(w, fmt.Sprintf("%d", vsess.UserID),
		vsess.Email, user.Email, vsess.Role,
		user.Role, user.FirstName, user.LastName, user.Password,
		fmt.Sprintf("%d", user.ID))
}

func FineUploadForEventGET(w http.ResponseWriter, r *http.Request) {
	_, vsess, _ := authenticatedSessionGet(w, r)
	// it is ok to ignore authentication errors, vsess can be nil for this function
	// check if user can acccess the files
	ev, err := GetEventFromFormCheckAccess(r, vsess)
	if err != nil {
		vlog.Warning.Printf("Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userID := ev.UserID
	editorID := uint32(0)
	if vsess != nil && vsess.Role == "editor" {
		userID = vsess.UserID
		editorID = vsess.UserID
	}
	x := struct {
		Name     string
		EventID  uint32
		UserID   uint32
		EditorID uint32
	}{Name: ev.Name, EventID: ev.ID, UserID: userID, EditorID: editorID}
	vviews.FineUploadMovies(w, x)
}

// func UploadMoviesALL(w http.ResponseWriter, r *http.Request) {
// 	s, vsess, err := authenticatedSessionGet(w, r)
// 	if err != nil {
// 		return
// 	}
// 	vlog.Trace.Printf("r.Method: %s\n", r.Method)
// 	if r.Method == "POST" {
// 		UploadMoviesPOST(w, r)
// 	} else {
// 		UploadMoviesGET(w, r)
// 	}
// }
// func UploadMoviesGET(w http.ResponseWriter, r *http.Request) {
// 	s, vsess, err := authenticatedSessionGet(w, r)
// 	if err != nil {
// 		return
// 	}
// 	vviews.UploadMovies(w)
// }
// func UploadMoviesPOST(w http.ResponseWriter, r *http.Request) {
// 	s, vsess, err := authenticatedSessionGet(w, r)
// 	if err != nil {
// 		return
// 	}
// 	//get the multipart reader for the request.
// 	reader, err := r.MultipartReader()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	if err = vfiles.SaveMultipart(vsess.UserID, reader); err != nil {
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
	_, _ = authenticatedSessionClear(w, r)
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
	user, err := vmodel.UserByEmail(email)
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
	_, err = authenticatedSessionSet(w, r, VSession{
		Authenticated: true,
		UserID:        user.ID,
		Email:         email,
		Role:          user.Role,
		FirstName:     user.FirstName,
		LastName:      user.LastName})
	if err != nil {
		return
	}
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
	err := vmodel.UserCreate(firstName, lastName, email, password, role)
	if err != nil {
		vlog.Warning.Printf("Create user %s returned err=%s", email, err.Error())
		// VG: show error page
		vviews.Error(w, "error or user already exits for email="+email+" error="+err.Error())
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func GetEventFromFormCheckAccess(r *http.Request, vsess *VSession) (ev vmodel.Event, err error) {
	eventID := r.FormValue("eventID")
	eid, err := stringToUint32(eventID)
	if err != nil {
		// vlog.Warning.Printf("eventID=%s is not integer, err=%v", eventID, err)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// check if the event belongs to this authenticated user
	ev, err = vmodel.EventGetByEventID(eid)
	if err != nil {
		// vlog.Warning.Printf("Could not find event id %d, err:%v", eid, err)
		// vviews.Error(w, "Could not find event id "+fmt.Sprintf("%d", eid)+" error="+err.Error())
		return
	}
	if vsess == nil {
		// if there is no session, the access is granted; the check for session should have taken place earlier
		return
	}
	allow, err := canAccessEvent(ev, vsess)
	if err != nil {
		return
	}
	if !allow {
		err = errors.New("User does not have permission to access event")
		return
	}
	return
}

func FilesGET(w http.ResponseWriter, r *http.Request) {
	_, vsess, err := authenticatedSessionGet(w, r)
	if err != nil {
		return
	}
	// check if user can acccess the files
	ev, err := GetEventFromFormCheckAccess(r, vsess)
	if err != nil {
		vlog.Warning.Printf("Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	RunDisplayFiles(w, ev, vsess)
}

func RunDisplayFiles(w http.ResponseWriter, ev vmodel.Event, vsess *VSession) {
	// get original files
	efs, err := vmodel.EventFileGetAllForEventID(ev.ID)
	if err != nil {
		vlog.Warning.Printf("Could not get files for events id %d, err:%v", ev.ID, err)
		vviews.Error(w, "Could not get events for user id "+fmt.Sprintf("%d", ev.ID)+" error="+err.Error())
		return
	}
	// get edited files
	var dfs []vmodel.EditedFile
	if vsess.Role == "editor" {
		dfs, err = vmodel.EditedFileGetAllForEventIDEditorID(ev.ID, vsess.UserID)
	} else {
		dfs, err = vmodel.EditedFileGetAllForEventID(ev.ID)
	}
	if err != nil {
		vlog.Warning.Printf("Could not get edited files for events id %d, editor %d err:%v", ev.ID, vsess.UserID, err)
		vviews.Error(w, "Could not get edited events for user id "+fmt.Sprintf("%d", ev.ID)+" error="+err.Error())
		return
	}
	vlog.Info.Printf("FielesShow: ev=%v, efs=%v, dfs=%v", ev, efs, dfs)
	vviews.FielesShow(w, ev, efs, dfs, vsess.Role)
}

func EventsGET(w http.ResponseWriter, r *http.Request) {
	_, vsess, err := authenticatedSessionGet(w, r)
	if err != nil {
		return
	}
	// get all events for this user
	evs, err := vmodel.EventGetAllForUserID(vsess.UserID)
	if err != nil {
		vlog.Warning.Printf("Could not get events for user id %d, err:%v", vsess.UserID, err)
		vviews.Error(w, "Could not get events for user id "+fmt.Sprintf("%d", vsess.UserID)+" error="+err.Error())
		return
	}
	vviews.EventsShow(w, evs)
}

func EditorEventsGET(w http.ResponseWriter, r *http.Request) {
	_, vsess, err := authenticatedSessionGet(w, r)
	if err != nil {
		return
	}
	// check that there is really an editor
	if vsess.Role != "editor" {
		vlog.Warning.Printf("user id %d is %s not editor", vsess.UserID, vsess.Role)
		vviews.Error(w, "Only editors can see editor events")
		return
	}
	// get all events for this editor
	ees, err := vmodel.EditorEventGetByEditorID(vsess.UserID)
	if err != nil {
		vlog.Warning.Printf("Could not get editor-events for editor with user id %d, err:%v", vsess.UserID, err)
		vviews.Error(w, "Could not get editor-events for editor with user id "+fmt.Sprintf("%d", vsess.UserID)+" error="+err.Error())
		return
	}
	type eeDesc struct {
		UserID        uint32
		UserNameFirst string
		UserNameLast  string
		UserEmail     string
		EventID       uint32
		EventName     string
		EventStatus   string
		CreatedAt     time.Time
		UpdatedAt     time.Time
	}
	list := make([]eeDesc, 0, len(ees))
	for _, ee := range ees {
		ev, err := vmodel.EventGetByEventID(ee.EventID)
		if err != nil {
			vlog.Warning.Printf("err on EventGetByEventID(%d) = %v", ee.EventID, err)
			continue
		}
		us, err := vmodel.UserGetByID(ev.UserID)
		if err != nil {
			vlog.Warning.Printf("err on UserGetByID(%d), evid=%d, err=%v", ev.UserID, ee.EventID, err)
			continue
		}
		x := eeDesc{
			UserID:        ev.UserID,
			UserNameFirst: us.FirstName,
			UserNameLast:  us.LastName,
			UserEmail:     us.Email,
			EventID:       ev.ID,
			EventName:     ev.Name,
			EventStatus:   ev.Status,
			CreatedAt:     ev.CreatedAt,
			UpdatedAt:     ev.UpdatedAt}
		list = append(list, x)
	}
	vviews.EditorEventsShow(w, list)
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
	_, vsess, err := authenticatedSessionGet(w, r)
	if err != nil {
		return
	}

	evName := r.FormValue("eventName")
	err = vmodel.EventCreate(vsess.UserID, evName)
	if err != nil {
		vlog.Warning.Printf("Create event for user id %d returned err=%s", vsess.UserID, err.Error())
		// VG: show error page
		vviews.Error(w, "Create event for user id "+fmt.Sprintf("%d", vsess.UserID)+" error="+err.Error())
		return
	}
	http.Redirect(w, r, "/events", http.StatusFound)
}

func UploadSignPOST(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		vlog.Warning.Printf("Method is not POST, but %s", r.Method)
		// VG: show error page
		vviews.Error(w, "Method is not POST, but "+r.Method)
		return
	}
	b, err := uploader.UploadFileCallbackBefore(r, vr, vs)
	if err != nil {
		vlog.Warning.Printf("uploader.UploadFileCallbackBefore err: %v", err)
		// VG: show error page
		vviews.Error(w, "uploader.UploadFileCallbackBefore err: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(b)
}

func UploadResultSuccess(w http.ResponseWriter, r *http.Request) {
	_, err := uploader.UploadFileCallbackAfter(r, vr, vs)
	if err != nil {
		vlog.Warning.Printf("uploader.UploadFileCallbackAfter err: %v", err)
		return
	}
	vlog.Trace.Printf("success!\n")
	//display success message.
	vviews.Hello(w, "upload", "successful")
}

func stringToUint32(s string) (n uint32, err error) {
	x, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		vlog.Warning.Printf("stringToUint32 s=%s, err=%v", s, err)
		return
	}
	n = uint32(x)
	return
}

func canAccessEvent(ev vmodel.Event, vsess *VSession) (bool, error) {
	if ev.UserID == vsess.UserID {
		return true, nil
	}
	// check if vsess.UserID is editor and if there is an association in editor_event
	if vsess.Role != "editor" {
		return false, nil
	}
	_, err := vmodel.EditorEventGetByEditorEventID(vsess.UserID, ev.ID)
	if err == nil {
		return true, nil
	}
	if err == vmodel.ErrNoResult {
		// this is not an error, but missing permission
		err = nil
	}
	return false, err
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
