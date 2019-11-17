package main

// Imports and initializations: {{{
import (
	"net/http"
	"html/template"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"log"
	"strings"
	"path"
	"net/url"
	"time"
)

type Item struct {
	Name string
	Path string
	Owner string
	Contents string
}

var templates = template.New("")
var sessionStore *sessions.FilesystemStore

//copied from previous work
func executeTemplate(w http.ResponseWriter, templ string, content interface{}) {
	err := templates.ExecuteTemplate(w, templ, content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var FILES Files
var DEFAULTSTYLE, DEFAULTSCRIPT, DEFAULTRENDEREDSTYLE []byte
// }}}

// Main: {{{
func main() {
	templates = template.Must(templates.Funcs(template.FuncMap{"base": path.Base, "dir": path.Dir}).ParseGlob("./views/*.html"))
	DBinit()

	key, err := ioutil.ReadFile("key")
	if err != nil {log.Fatal(err)}
	sessionStore = sessions.NewFilesystemStore("", key)

	FILES = &FSFiles{}
	FILES.Init("./files")
	DEFAULTSTYLE, err = ioutil.ReadFile("views/default.css")
	DEFAULTSCRIPT, err = ioutil.ReadFile("views/default.js")
	DEFAULTRENDEREDSTYLE, err = ioutil.ReadFile("views/rendered.css")
	if err != nil {log.Fatal(err)}

	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/newUser", newUser)

	http.HandleFunc("/new/", reqAuth(newFile))
	http.HandleFunc("/newFolder/", reqAuth(newFolder))

	http.HandleFunc("/css.css", reqAuth(serveCss))
	http.HandleFunc("/userScript.js", reqAuth(serveJs))
	http.HandleFunc("/rendered.css", serveRendered)
	http.HandleFunc("/file/", reqAuth(showFile))
	http.HandleFunc("/edit/", reqAuth(editFile))
	http.HandleFunc("/upload/", reqAuth(upload))
	http.HandleFunc("/renameFolder/", reqAuth(renameFolder))
	http.HandleFunc("/remove/", reqAuth(remove))
	http.HandleFunc("/render/", reqAuth(render))

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
		index, err := FILES.Get(username, "/")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		executeTemplate(w, "index.html", struct{Logged bool; File File}{true, index})
		return
	}
	http.NotFound(w, r)
}
// }}}

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
func reqAuth(handler func(http.ResponseWriter, *http.Request, string)) func(http.ResponseWriter, *http.Request) {
	return func (w http.ResponseWriter, r *http.Request) {
		username, err := authUser(w, r)
		if err != nil {return}
		handler(w, r, username)
	}
}

// }}}

// Create Methods: {{{
// !!!NOTE!!! - Variable "path" in some of these functions clashes with "path" library---that may cause a bug later
func getFilePathFromURL(u *url.URL, prefix string) (string, error) {
	path := u.Path[len(prefix):]
	var err error
	if u.RawQuery != "" {
		var rq string
		rq, err = url.QueryUnescape(u.RawQuery)
		path += "?" + rq
	}
	if u.Fragment != "" {
		path += "#" + u.Fragment
	}
	return path, err
}

func newFile(w http.ResponseWriter, r *http.Request, owner string) {
	name, err := getFilePathFromURL(r.URL, "/new/")
	if err != nil {
		http.Error(w, "Bad Filename", http.StatusBadRequest)
		return
	}
	if name == "" {
		name = "Untitled"
	} else {
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
	http.Redirect(w, r, "/file/" + file.Path + "#", http.StatusFound)
}
func newFolder(w http.ResponseWriter, r *http.Request, owner string) {
	name, err := getFilePathFromURL(r.URL, "/newFolder/")
	if err != nil {
		http.Error(w, "Bad Filename", http.StatusBadRequest)
		return
	}
	if name == "" {
		name = "Untitled"
	} else {
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
func upload(w http.ResponseWriter, r *http.Request, owner string) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	// one gibibyte upload limit
	r.ParseMultipartForm(2 << 30)
	file, info, err := r.FormFile("upload")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	valid, errStr := validate(info.Filename)
	if !valid {
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}
	filepath, err := getFilePathFromURL(r.URL, "/upload/")
	if err != nil {
		http.Error(w, "Bad Filename", http.StatusBadRequest)
		return
	}
	fullfilepath := filepath + "/" + info.Filename
	defer file.Close()
	filecontents, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	FILES.Edit(owner, fullfilepath, string(filecontents))
	http.Redirect(w, r, "/file/" + filepath, http.StatusFound)
}
//consider changing to just string & chekcing if it is empty
func validate(name string) (bool, string) { //valid/not, error
	forbiddenStrings := []string{
		"/", "\x00", /* null byte */
		"#", /* reference fragments can go fuck themselves */
		"%", /* deciding between ? and %, I had to allow ?... */
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

// Show files: {{{
// !!!NOTE!!! - Variable "path" in some of these functions clashes with "path" library---that may cause a bug later
func serveCss(w http.ResponseWriter, r *http.Request, owner string) {
	if r.URL.Path != "/css.css" {
		http.NotFound(w, r)
		return
	}
	var style []byte
	if owner == "" {
		style = DEFAULTSTYLE
	} else {
		styleFile, err := FILES.Get(owner, "/.style.css")
		if err != nil || styleFile.Filetype == FOLDER {
			style = DEFAULTSTYLE
		} else {
			style = []byte(styleFile.FileContents)
		}
	}
	w.Header().Set("Content-Type", "text/css")
	w.WriteHeader(200)
	w.Write(style)
}
func serveJs(w http.ResponseWriter, r *http.Request, owner string) {
	if r.URL.Path != "/userScript.js" {
		http.NotFound(w, r)
		return
	}
	if owner == "" {return} //shouldn't happen
	var script []byte
	jsFile, err := FILES.Get(owner, "/.userScript.js")
	if err != nil || jsFile.Filetype == FOLDER {
		script = DEFAULTSCRIPT
	} else {
		script = []byte(jsFile.FileContents)
	}
	w.Header().Set("Content-Type", "text/javascript")
	w.WriteHeader(200)
	w.Write(script)
}
func showFile(w http.ResponseWriter, r *http.Request, owner string) {
	if owner == "" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	path, err := getFilePathFromURL(r.URL, "/file/")
	if err != nil {
		http.Error(w, "Bad Filename", http.StatusBadRequest)
		return
	}
	if path == "" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	file, err := FILES.Get(owner, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	switch file.Filetype {
	case FILE:
		executeTemplate(w, "file.html", file)
	case FOLDER:
		executeTemplate(w, "folder.html", file)
	}
}
// }}}

// Edit/rename/remove files: {{{zo
func editFile(w http.ResponseWriter, r *http.Request, owner string) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	valid, errStr := validate(path.Base(r.FormValue("title")))
	if !valid {
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}
	filepath, err := getFilePathFromURL(r.URL, "/edit/")
	if err != nil {
		http.Error(w, "Bad Filename", http.StatusBadRequest)
		return
	}
	if filepath == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	//the following bans people from editing a file that doesn't exist
	//should I allow it? That way, people could have a button to create & autofill a file with stuff
	//TODO: think about this
	file, err := FILES.Get(owner, filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if file.Filetype != FILE {
		http.Error(w, "You are trying to edit a folder instead of a file!", http.StatusBadRequest)
		return
	}
	err = FILES.Edit(owner, filepath, r.FormValue("filecontents"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if r.FormValue("title") != "" && r.FormValue("title") != path.Base(file.Path) {
		err = FILES.Rename(owner, filepath, path.Dir(filepath) + "/" + r.FormValue("title"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/file/" + path.Dir(filepath) + "/" + r.FormValue("title") + "#", http.StatusSeeOther)
		return
	}
	//TODO: actual content indicator. THis is temporary to keep them on the page while we don't use JS
	http.Redirect(w, r, "/file/" + filepath + "#", http.StatusSeeOther)
}
func renameFolder(w http.ResponseWriter, r *http.Request, owner string) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	valid, errStr := validate(path.Base(r.FormValue("name")))
	if !valid {
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}
	if owner == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	filepath, err := getFilePathFromURL(r.URL, "/renameFolder/")
	if err != nil {
		http.Error(w, "Bad Filename:" + err.Error(), http.StatusBadRequest)
		return
	}
	if filepath == "" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	folder, err := FILES.Get(owner, filepath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if folder.Filetype != FOLDER {
		http.Error(w, "You are trying to rename a file instead of a folder. To rename a file, open the file, change its title, and press \"Save\"", http.StatusBadRequest)
		return
	}
	if r.FormValue("name") != "" && r.FormValue("name") != path.Base(folder.Path) {
		err = FILES.Rename(owner, filepath, path.Dir(filepath) + "/" + r.FormValue("name"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/file/" + path.Dir(filepath) + "/" + r.FormValue("name") + "#", http.StatusSeeOther)
		return
	}
	http.Error(w, "No folder name provided or folder name has not changed", http.StatusBadRequest)
}

func remove(w http.ResponseWriter, r *http.Request, owner string) {
	filename, err := getFilePathFromURL(r.URL, "/remove/")
	if err != nil {
		http.Error(w, "Bad Filename", http.StatusBadRequest)
		return
	}
	err = FILES.Remove(owner, filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/file/" + path.Dir(filename) + "#", http.StatusFound)
}
// }}}

// Render files: {{{
func render(w http.ResponseWriter, r *http.Request, owner string) {
	pathChunks := strings.Split(r.URL.Path[len("/render/"):], "/")
	rednerer, exists := RENDERFUNCS[pathChunks[0]]
	if !exists && pathChunks[0] != "plain" {
		http.Error(w, "Renderer not found", http.StatusNotFound)
		return
	}
	filename, err := getFilePathFromURL(r.URL, "/render/" + pathChunks[0] + "/")
	if err != nil {
		http.Error(w, "Bad Filename", http.StatusBadRequest)
		return
	}
	if filename == "" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
	file, err := FILES.Get(owner, filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if file.Filetype == FOLDER {
		http.Error(w, "You cannot render a folder", http.StatusBadRequest)
		return
	}
	if pathChunks[0] == "plain" {
		http.ServeContent(w, r, filename, time.Now(), strings.NewReader(file.FileContents))
		return
	}
	renderedText, err := rednerer(file.FileContents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	executeTemplate(w, "rendered.html", Rendered{filename, renderedText})
}

func serveRendered(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/rendered.css" {
		http.NotFound(w, r)
		return
	}
	var style []byte
	owner, err := authUser(w, r)
	if err != nil {return}
	if owner == "" {
		style = DEFAULTRENDEREDSTYLE
	} else {
		styleFile, err := FILES.Get(owner, "/.renderedStyle.css")
		if err != nil || styleFile.Filetype == FOLDER {
			style = DEFAULTRENDEREDSTYLE
		} else {
			style = []byte(styleFile.FileContents)
		}
	}
	w.Header().Set("Content-Type", "text/css")
	w.WriteHeader(200)
	w.Write(style)
}
// }}}
