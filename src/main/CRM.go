package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	//"github.com/gorilla/sessions"
)

type customer_struct struct {
	customer_name string
	customer_id   string
	customer_type string
}

var customer_map = make(map[string]customer_struct)

var cookiemap = make(map[string]string)
var users = make(map[string]string)

const cookieName = "CookieCRM"

type ViewData struct {
	Title     string
	Message   string
	User      string
	Customers map[string]customer_struct
}

func GenerateId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
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

	nameUserFromCookieStruc := ""

	CookieGet, _ := r.Cookie(cookieName)
	if CookieGet != nil {
		nameUserFromCookie, flagmap := cookiemap[CookieGet.Value]
		if flagmap != false {
			nameUserFromCookieStruc = nameUserFromCookie
		}
	}

	data := ViewData{
		Title:     "list customer",
		Message:   "list customer below",
		User:      nameUserFromCookieStruc,
		Customers: customer_map,
	}

	// t.ExecuteTemplate(w, "main_page", customer_map)
	t.ExecuteTemplate(w, "main_page", data)
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

// //examples
// func users(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFile(w, r, "user.html")
// }

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
		User:    "admin",
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

func login(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./login/login.html")
}

func loginPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user_password, flagusers := users[username]
	if flagusers == true {
		if user_password != password {
			fmt.Fprint(w, "error auth password")
			return
		}
	} else {
		fmt.Fprint(w, "error auth user not find")
		return
	}

	idcookie := GenerateId()

	cookiemap[idcookie] = username

	cookieHttp := &http.Cookie{
		Name:    cookieName,
		Value:   idcookie,
		Expires: time.Now().Add(6 * time.Minute),
	}

	http.SetCookie(w, cookieHttp)

	//fmt.Fprint(w, username+" "+password)
	http.Redirect(w, r, "/", 302)
}

func main() {

	users["admin"] = "admin"

	router := mux.NewRouter()

	router.HandleFunc("/", indexHandler)

	//localhost:8181/products/2
	router.HandleFunc("/products/{id:[0-9]+}", productsHandler)

	//localhost:8181/user?name=Sam&age=21
	router.HandleFunc("/user", user)

	//localhost:8181/get_customer?customer_id="123"
	router.HandleFunc("/get_customer", get_customer)

	// router.HandleFunc("/users", users)
	router.HandleFunc("/postform", postform)

	router.HandleFunc("/add_change_customer", add_change_customer)
	router.HandleFunc("/postform_add_change_customer", postform_add_change_customer)

	router.HandleFunc("/templates", templates)
	router.HandleFunc("/list_customer", list_customer)

	router.HandleFunc("/mainpage", mainpage)

	router.HandleFunc("/login", login)
	router.HandleFunc("/loginPost", loginPost)

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
