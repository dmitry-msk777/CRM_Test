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

type customer_struct struct {
	customer_name string
	customer_id   string
	customer_type string
}

var customer_map = make(map[string]customer_struct)

var cookieStore = sessions.NewCookieStore([]byte("secret"))

const cookieName = "MyCookie"

type sesKey int

type ViewData struct {
	Title     string
	Message   string
	Customers map[string]customer_struct
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	response := fmt.Sprintf("id=%s", id)
	fmt.Fprint(w, response)
}
func indexHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("templates/main_page.html", "templates/header.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	t.ExecuteTemplate(w, "main_page", customer_map)
}

//examples
func user(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	age := r.URL.Query().Get("age")
	fmt.Fprintf(w, "Имя: %s Возраст: %s", name, age)
}

func get_customer(w http.ResponseWriter, r *http.Request) {
	customer_id := r.URL.Query().Get("customer_id")
	fmt.Fprintf(w, "customer_id: %s customer_name: %s", customer_map[customer_id].customer_id,
		customer_map[customer_id].customer_name)
}

//examples
func users(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "user.html")
}

//examples
func postform(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("username")
	age := r.FormValue("userage")
	fmt.Fprintf(w, "Имя: %s Возраст: %s", name, age)
}

func add_change_customer(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/add_change_customer.html", "templates/header.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl.ExecuteTemplate(w, "add_change_customer", nil)

}

func postform_add_change_customer(w http.ResponseWriter, r *http.Request) {
	customer_data := customer_struct{
		customer_name: r.FormValue("customer_name"),
		customer_id:   r.FormValue("customer_id"),
		customer_type: r.FormValue("customer_type"),
	}

	customer_map[r.FormValue("customer_id")] = customer_data

	//fmt.Fprintf(w, "Имя: %s Идентификатор: %s", customer_data.customer_name, customer_data.customer_id)

	http.Redirect(w, r, "/", 302)
}

//examples
func templates(w http.ResponseWriter, r *http.Request) {
	data := ViewData{
		Title:   "World Cup",
		Message: "FIFA will never regret it",
	}
	tmpl, _ := template.ParseFiles("templates/templates.html")
	tmpl.Execute(w, data)
}

func list_customer(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/list_customer.html", "templates/header.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	data := ViewData{
		Title:     "list customer",
		Message:   "list customer below",
		Customers: customer_map,
	}

	tmpl.ExecuteTemplate(w, "list_customer", data)
}

func mainpage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Index Page")
}

func login_old(w http.ResponseWriter, r *http.Request) {
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

func login(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./login/login.html")
}

func loginPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	fmt.Fprint(w, username+" "+password)
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

	//localhost:8181/get_customer?customer_id="123"
	router.HandleFunc("/get_customer", get_customer)

	router.HandleFunc("/users", users)
	router.HandleFunc("/postform", postform)

	router.HandleFunc("/add_change_customer", add_change_customer)
	router.HandleFunc("/postform_add_change_customer", postform_add_change_customer)

	router.HandleFunc("/templates", templates)
	router.HandleFunc("/list_customer", list_customer)

	router.HandleFunc("/mainpage", mainpage)

	router.HandleFunc("/login", login)
	router.HandleFunc("/loginPost", loginPost)

	router.HandleFunc("/loginind", loginind)

	router.HandleFunc("/logintest", logintest)

	var dir string
	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()

	//router.Handle("/js/", http.FileServer(http.Dir("./js/")))
	//Работает
	router.PathPrefix("/js").Handler(http.StripPrefix("/js", http.FileServer(http.Dir("./js/"))))

	http.Handle("/", router)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8181", nil)
}
