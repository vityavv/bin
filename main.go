package main

import (
	"net/http"
	"html/template"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"log"
	"strings"
	"path"
)

type Item struct {
	Name string
	Path string
	Owner string
	Contents string
}

var templates = template.Must(template.ParseGlob("./views/*.html"))
var sessionStore *sessions.FilesystemStore

//copied from previous work
func executeTemplate(w http.ResponseWriter, templ string, content interface{}) {
	err := templates.ExecuteTemplate(w, templ, content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var FILES Files

func main() {
	DBinit()

	key, err := ioutil.ReadFile("key")
	if err != nil {log.Fatal(err)}
	sessionStore = sessions.NewFilesystemStore("", key)

	FILES = &FSFiles{}
	FILES.Init("./files")

	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/newUser", newUser)

	http.HandleFunc("/new/", newFile)
	http.HandleFunc("/newFolder/", newFolder)

	http.HandleFunc("/", mainHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/index" {
		username, err := authUser(w, r)
		if err != nil {return}
		if username == "" {
			executeTemplate(w, "index.html", struct{Logged bool}{false})
			return
		}
		//get stuff TODO
		executeTemplate(w, "index.html", struct{Logged bool; Items []Item}{true, []Item{Item{"Test", "/test", "everyone", "stuff"}}})
		return
	}
	http.NotFound(w, r)
}

// User Methods: {{{
func newUser(w http.ResponseWriter, r *http.Request) {
	username, err := authUser(w, r)
	if err != nil {return}
	if username != "" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	switch r.Method {
	case "GET":
		executeTemplate(w, "newUser.html", "")
	case "POST":
		if r.FormValue("username") == "" || r.FormValue("password") == "" || r.FormValue("password2") == "" {
			http.Error(w, "You are missing one of the fields!", http.StatusBadRequest)
			return
		}
		if r.FormValue("password") != r.FormValue("password2") {
			http.Error(w, "The passwords do not match!", http.StatusBadRequest)
			return
		}
		err := DBcreateUser(r.FormValue("username"), r.FormValue("password"))
		if err != nil {
			http.Error(w, "Error: " + err.Error(), http.StatusInternalServerError)
			return
		}
		session, err := sessionStore.Get(r, "user")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		session.Values["username"] = r.FormValue("username")
		session.Values["password"] = r.FormValue("password")
		session.Save(r, w)
		err = FILES.NewUser(r.FormValue("username"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, "user")
	if err != nil {
		http.Error(w, "Internal Server Error: " + err.Error(), http.StatusInternalServerError)
		return
	}
	delete(session.Values, "username")
	delete(session.Values, "password")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}
func login(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.NotFound(w, r)
	}
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	username, err := authUser(w, r)
	if username != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if err != nil {
		return //error already handled by function
	}
	if r.FormValue("username") == "" || r.FormValue("password") == "" {
		http.Error(w, "Either the username or the password is missing", http.StatusBadRequest)
		return
	}
	loggedin, err := DBlogIn(r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		http.Error(w, "Error: " + err.Error(), http.StatusInternalServerError)
		return
	}
	if !loggedin {
		http.Error(w, "Password is incorrect", http.StatusUnauthorized)
		return
	}
	session, err := sessionStore.Get(r, "user")
	if err != nil {
		http.Error(w, "Internal Server Error: " + err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["username"] = r.FormValue("username")
	session.Values["password"] = r.FormValue("password")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func authUser(w http.ResponseWriter, r *http.Request) (string, error) { // username, error, status code
	session, err := sessionStore.Get(r, "user")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return "", err
	}
	interface_username, exists := session.Values["username"]
	interface_password, exists := session.Values["password"]
	if !exists {
		return "", nil
	}
	username, password := interface_username.(string), interface_password.(string) //TODO: type switch + error
	loggedIn, err := DBlogIn(username, password)
	if err != nil {
		http.Error(w, "Error logging in", http.StatusInternalServerError)
		return "", err
	}
	if !loggedIn {
		return "", nil
	}
	return username, nil
}
// }}}

// File Methods: {{{
func newFile(w http.ResponseWriter, r *http.Request) {
	owner, err := authUser(w, r)
	if err != nil {
		return
	}
	var name string
	if len(r.URL.Path) <= len("/new/") {
		name = "Untitled"
	} else {
		name = r.URL.Path[len("/new/"):]
		exists, err := FILES.Get(owner, name)
		if err == nil {
			if exists.Filetype == FOLDER {
				name = name + "/Untitled"
			} else {
				http.Error(w, "File already exists!", http.StatusBadRequest)
				return
			}
		}
		valid, errStr := validate(path.Base(name))
		if !valid {
			http.Error(w, errStr, http.StatusBadRequest)
		}
	}
	file, err := FILES.New(owner, name)
	if err != nil {
		//change so that it sees what the error is & acts on that (w/ type switch?)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/file/" + file.Path, http.StatusFound)
}
func newFolder(w http.ResponseWriter, r *http.Request) {
	owner, err := authUser(w, r)
	if err != nil {
		return
	}
	var name string
	if len(r.URL.Path) <= len("/newFolder/") {
		name = "Untitled"
	} else {
		name = r.URL.Path[len("/newFolder/"):]
		exists, err := FILES.Get(owner, name)
		if err == nil {
			if exists.Filetype == FOLDER {
				name = name + "/Untitled"
			} else {
				http.Error(w, "File already exists!", http.StatusBadRequest)
				return
			}
		}
		valid, errStr := validate(path.Base(name))
		if !valid {
			http.Error(w, errStr, http.StatusBadRequest)
		}
	}
	folder, err := FILES.NewFolder(owner, string(name))
	if err != nil {
		//change so that it sees what the error is & acts on that (w/ type switch?)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/file/" + folder.Path, http.StatusFound)
}
//consider changing to just string & chekcing if it is empty
func validate(name string) (bool, string) { //valid/not, error
	forbiddenStrings := []string{
		"/", "\x00", /* null byte */
	}
	for _, s := range forbiddenStrings {
		if strings.Contains(name, s) {
			return false, "Improper filename; filename cannot have " + s
		}
	}
	if name == "." || name == ".." {
		return false, "Improper filename; filename cannot be . or .."
	}
	return true, ""
}
// }}}
