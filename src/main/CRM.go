package main

import (
	"fmt"
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
)

type ViewData struct{
    Title string
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
		Title: "World Cup",
		Message: "FIFA will never regret it",
	}
	tmpl, _ := template.ParseFiles("templates/templates.html")
	tmpl.Execute(w, data)
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/", indexHandler)

	//localhost:8181/products/2
	router.HandleFunc("/products/{id:[0-9]+}", productsHandler)

	//localhost:8181/user?name=Sam&age=21
	router.HandleFunc("/user", user)

	router.HandleFunc("/users", users)
	router.HandleFunc("/postform", postform)

	router.HandleFunc("/templates", templates)


	http.Handle("/", router)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8181", nil)
}
