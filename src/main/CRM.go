package main

import (
	"crypto/rand"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	//"github.com/gorilla/sessions"
	"encoding/base64"
	"net/mail"
	"net/smtp"
)

type customer_struct struct {
	customer_name  string
	customer_id    string
	customer_type  string
	customer_email string
}

var customer_map = make(map[string]customer_struct)

var cookiemap = make(map[string]string)
var users = make(map[string]string)

var mass_settings = make([]string, 2)

var type_memory_storage string

const cookieName = "CookieCRM"

type ViewData struct {
	Title     string
	Message   string
	User      string
	Customers map[string]customer_struct
}

type GetINNResponse struct {
	INN string `xml:"INN"`
}

var resINN GetINNResponse

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

func encodeRFC2047(String string) string {
	// use mail's rfc2047 to encode any string
	addr := mail.Address{String, ""}
	return strings.Trim(addr.String(), " <>")
}

func postform_add_change_customer(w http.ResponseWriter, r *http.Request) {
	customer_data := customer_struct{
		customer_name:  r.FormValue("customer_name"),
		customer_id:    r.FormValue("customer_id"),
		customer_type:  r.FormValue("customer_type"),
		customer_email: r.FormValue("customer_email"),
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

func email_settings(w http.ResponseWriter, r *http.Request) {
	// Add fill elements form from a global variable or database
	// Add the ability to select an smtp-server or extract a server from an email address
	http.ServeFile(w, r, "./mail_smtp/settings.html")
}

func email_settingsPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	//fmt.Fprint(w, email+"error auth user not find"+password)

	mass_settings[0] = email
	mass_settings[1] = password

	http.Redirect(w, r, "/", 302)
}

func send_message(w http.ResponseWriter, r *http.Request) {

	// Set up authentication information. https://yandex.ru/support/mail/mail-clients.html

	smtpServer := "smtp.yandex.ru"
	auth := smtp.PlainAuth(
		"",
		mass_settings[0],
		mass_settings[1],
		smtpServer,
	)

	from := mail.Address{"Test", mass_settings[0]}
	to := mail.Address{"test2", "dima-irk35@mail.ru"}
	title := "Title"

	body := "body"

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = encodeRFC2047(title)
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	err := smtp.SendMail(
		smtpServer+":25",
		auth,
		from.Address,
		[]string{to.Address},
		[]byte(message),
		//[]byte("This is the email body."),
	)
	if err != nil {
		fmt.Fprint(w, "error"+err.Error())
	} else {
		http.Redirect(w, r, "/", 302)
	}

}

func soap_get(w http.ResponseWriter, r *http.Request) {
	// ////Work with SOAP "github.com/tiaguinho/gosoap"
	// // Do not job check site https://infostart.ru/public/439808/
	// soap, err := gosoap.SoapClient("http://npchk.nalog.ru/FNSNDSCAWS_2?wsdl")
	// if err != nil {
	// 	fmt.Errorf("error not expected: %s", err)
	// }

	// params := gosoap.Params{
	// 	"INN": "7702807750",
	// }

	// err = soap.Call("FNSNDSCAWS2", params)
	// if err != nil {
	// 	fmt.Errorf("error in soap call: %s", err)
	// }

	// soap.Unmarshal(&resINN)
	// // if r.GetINNResponse.CountryCode != "USA" {
	// // 	fmt.Errorf("error: %+v", r)
	// // }
	// fmt.Println(resINN)
}

func main() {

	//split handler so
	//if r.Method == "POST" {

	type_memory_storage_flag := flag.String("type_memory_storage", "", "type storage data")
	flag.Parse()

	if *type_memory_storage_flag == "" {
		type_memory_storage = "global_variable"
	} else {
		type_memory_storage = *type_memory_storage_flag
	}

	//temporary
	*type_memory_storage_flag = "SQLit"

	if *type_memory_storage_flag == "SQLit" {
		db, err := sql.Open("sqlite3", "./bd/SQLit/base_sqlit.db")

		if err != nil {
			panic(err)
		}
		defer db.Close()
		//result, err := db.Exec("insert into products (model, company, price) values ('iPhone X', $1, $2)",
		//    "Apple", 72000)
		result, err := db.Exec("insert into customer (id, email) values ('123', $1)",
			"Apple")
		if err != nil {
			panic(err)
		}
		fmt.Println(result.LastInsertId()) // id последнего добавленного объекта
		fmt.Println(result.RowsAffected()) // количество добавленных строк

	}

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

	router.HandleFunc("/email_settings", email_settings)
	router.HandleFunc("/email_settingsPost", email_settingsPost)

	router.HandleFunc("/send_message", send_message)
	router.HandleFunc("/soap_get", soap_get)

	// var dir string
	// flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	// flag.Parse()

	//router.Handle("/js/", http.FileServer(http.Dir("./js/")))
	//Работает
	router.PathPrefix("/js").Handler(http.StripPrefix("/js", http.FileServer(http.Dir("./js/"))))

	http.Handle("/", router)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":8181", nil)

}
