package main

import (
	"net/http"
	"html/template"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"log"
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

func main() {
	key, err := ioutil.ReadFile("key")
	if err != nil {log.Fatal(err)}
	sessionStore = sessions.NewFilesystemStore("", key)

	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	//http.HandleFunc("/new/", newFile)
	http.HandleFunc("/", mainHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/index" {
		username, err := authUser(w, r)
		if err != nil {
			return
		}
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
	//loggedin, err := DBlogIn(r.FormValue("username"), r.FormValue("password"))
	loggedin := true //TODO
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
	log.Println(password)
	//loggedIn, err := DBlogIn(username, password)
	loggedIn := true //TODO
	if err != nil {
		http.Error(w, "Error logging in", http.StatusInternalServerError)
		return "", err
	}
	if !loggedIn {
		return "", nil
	}
	return username, nil
}
