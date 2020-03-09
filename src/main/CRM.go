package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

const (
	sesKeyLogin sesKey = iota
)

var cookieStore = sessions.NewCookieStore([]byte("secret"))

const cookieName = "MyCookie"

type sesKey int

type ViewData struct {
	Title   string
	Message string
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	response := fmt.Sprintf("id=%s", id)
	fmt.Fprint(w, response)
}
func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Index Page")
}

func user(w http.ResponseWriter, r *http.Request) {

	name := r.URL.Query().Get("name")
	age := r.URL.Query().Get("age")
	fmt.Fprintf(w, "Имя: %s Возраст: %s", name, age)
}

func users(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "user.html")
}

func postform(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("username")
	age := r.FormValue("userage")

	fmt.Fprintf(w, "Имя: %s Возраст: %s", name, age)
}

func templates(w http.ResponseWriter, r *http.Request) {

	data := ViewData{
		Title:   "World Cup",
		Message: "FIFA will never regret it",
	}
	tmpl, _ := template.ParseFiles("templates/templates.html")
	tmpl.Execute(w, data)
}

func mainpage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Index Page")
}

func login(w http.ResponseWriter, r *http.Request) {
	ses, err := cookieStore.Get(r, cookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ses.Values[sesKeyLogin] = "user"
	err = cookieStore.Save(r, w, ses)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func loginind(w http.ResponseWriter, r *http.Request) {
	ses, err := cookieStore.Get(r, cookieName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	login, ok := ses.Values[sesKeyLogin].(string)
	if !ok {
		login = "anonymous"
	}

	w.Write([]byte("you are " + login))
}

func logintest(w http.ResponseWriter, r *http.Request) {
	url := "login/login.html" //Страница входа
	if len(r.Header["Cookie"]) != 0 && r.Header["Cookie"][0] == "auth=your_MD5_cookies" {
		url = "login/index.html" //Страница после успешной авторизации
	}
	t, _ := template.ParseFiles(url)
	t.Execute(w, "")
}

func main() {

	gob.Register(sesKey(0))

	router := mux.NewRouter()

	router.HandleFunc("/", indexHandler)

	//localhost:8181/products/2
	router.HandleFunc("/products/{id:[0-9]+}", productsHandler)

	//localhost:8181/user?name=Sam&age=21
	router.HandleFunc("/user", user)

	router.HandleFunc("/users", users)
	router.HandleFunc("/postform", postform)

	router.HandleFunc("/templates", templates)

	router.HandleFunc("/mainpage", mainpage)

	router.HandleFunc("/login", login)

	router.HandleFunc("/loginind", loginind)

	router.HandleFunc("/logintest", logintest)

	var dir string
	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()

	//router.Handle("/js/", http.FileServer(http.Dir("./js/")))
	//router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("/login/js/"))))
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./js/"))))

	http.Handle("/", router)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8181", nil)
}
