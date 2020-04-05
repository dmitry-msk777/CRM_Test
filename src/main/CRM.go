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

	"log"
	"os"

	"github.com/gorilla/mux"
	//"github.com/gorilla/sessions"
	"encoding/base64"
	"net/mail"
	"net/smtp"

	//"github.com/tiaguinho/gosoap"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"

	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

type Customer_struct struct {
	Customer_id    string
	Customer_name  string
	Customer_type  string
	Customer_email string
}

type users_base struct {
	user     string
	password string
}

type cookie_base struct {
	id   string
	user string
}

type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	S       string   `xml:"S,attr"`
	Body    struct {
		Text         string `xml:",chardata"`
		NdsResponse2 struct {
			Text    string `xml:",chardata"`
			Xmlns   string `xml:"xmlns,attr"`
			DTActFL string `xml:"DTActFL,attr"`
			DTActUL string `xml:"DTActUL,attr"`
			NP      struct {
				Text  string `xml:",chardata"`
				INN   string `xml:"INN,attr"`
				State string `xml:"State,attr"`
			} `xml:"NP"`
		} `xml:"NdsResponse2"`
	} `xml:"Body"`
}

var database *sql.DB

var customer_map = make(map[string]Customer_struct)

var cookiemap = make(map[string]string)
var users = make(map[string]string)

var mass_settings = make([]string, 2)

var type_memory_storage string

const cookieName = "CookieCRM"

type ViewData struct {
	Title     string
	Message   string
	User      string
	Customers map[string]Customer_struct
}

var InfoLogger *log.Logger
var ErrorLogger *log.Logger

func GenerateId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("templates/main_page.html", "templates/header.html")
	if err != nil {
		ErrorLogger.Println(err.Error())
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

	if type_memory_storage == "SQLit" && CookieGet != nil {

		rows, err := database.Query("select * from cookie where id = $1", CookieGet.Value)
		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		defer rows.Close()
		cookie_base_s := []cookie_base{}

		for rows.Next() {
			p := cookie_base{}
			err := rows.Scan(&p.id, &p.user)
			if err != nil {
				ErrorLogger.Println(err.Error())
				fmt.Println(err)
				continue
			}
			cookie_base_s = append(cookie_base_s, p)
		}
		for _, p := range cookie_base_s {
			nameUserFromCookieStruc = p.user
			fmt.Println(p.id, p.user)
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
	fmt.Fprintf(w, "customer_id: %s customer_name: %s", customer_map[customer_id].Customer_id,
		customer_map[customer_id].Customer_name)
}

func add_change_customer(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/add_change_customer.html", "templates/header.html")
	if err != nil {
		ErrorLogger.Println(err.Error())
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
	customer_data := Customer_struct{
		Customer_name:  r.FormValue("customer_name"),
		Customer_id:    r.FormValue("customer_id"),
		Customer_type:  r.FormValue("customer_type"),
		Customer_email: r.FormValue("customer_email"),
	}

	if type_memory_storage == "SQLit" {

		_, err := database.Exec("insert into customer (customer_id, customer_name, customer_type, customer_email) values (?, ?, ?, ?)",
			customer_data.Customer_id, customer_data.Customer_name, customer_data.Customer_type, customer_data.Customer_email)

		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}
		http.Redirect(w, r, "list_customer", 301)

	} else {
		customer_map[r.FormValue("customer_id")] = customer_data
	}

	http.Redirect(w, r, "/list_customer", 302)
}

func list_customer(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/list_customer.html", "templates/header.html")
	if err != nil {
		ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	if type_memory_storage == "SQLit" {

		var customer_map_s = make(map[string]Customer_struct)

		rows, err := database.Query("select * from customer")
		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		defer rows.Close()
		Customer_struct_s := []Customer_struct{}

		for rows.Next() {
			p := Customer_struct{}
			err := rows.Scan(&p.Customer_id, &p.Customer_name, &p.Customer_type, &p.Customer_email)
			if err != nil {
				ErrorLogger.Println(err.Error())
				fmt.Println(err)
				continue
			}
			Customer_struct_s = append(Customer_struct_s, p)
		}
		for _, p := range Customer_struct_s {
			customer_map_s[p.Customer_id] = p
		}

		tmpl.ExecuteTemplate(w, "list_customer", customer_map_s)
	} else {
		tmpl.ExecuteTemplate(w, "list_customer", customer_map)
	}

	// data := ViewData{
	// 	Title:     "list customer",
	// 	Message:   "list customer below",
	// 	Customers: customer_map,
	// }

	//tmpl.ExecuteTemplate(w, "list_customer", data)

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

	if type_memory_storage == "SQLit" {

		rows, err := database.Query("select * from users where user = $1 and password = $2", username, password)
		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		defer rows.Close()
		users_base_s := []users_base{}

		for rows.Next() {
			p := users_base{}
			err := rows.Scan(&p.user, &p.password)
			if err != nil {
				ErrorLogger.Println(err.Error())
				fmt.Println(err)
				continue
			}
			users_base_s = append(users_base_s, p)
		}
		for _, p := range users_base_s {
			fmt.Println(p.user, p.password)
		}

	} else {

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
	}

	idcookie := GenerateId()

	if type_memory_storage == "SQLit" {

		result, err := database.Exec("insert into cookie (id, user) values ($1, $2)",
			idcookie, username)
		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		fmt.Println(result.LastInsertId()) // id последнего добавленного объекта
		fmt.Println(result.RowsAffected()) // количество добавленных строк

	} else {
		cookiemap[idcookie] = username
	}

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

	if r.Method == "GET" {

		tmpl, err := template.ParseFiles("mail_smtp/settings.html", "templates/header.html")
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		data := ViewData{
			Title:     "Test777",
			Message:   mass_settings[1],
			User:      mass_settings[0],
			Customers: nil,
		}

		if type_memory_storage == "SQLit" {

			// var customer_map_s = make(map[string]Customer_struct)

			// rows, err := database.Query("select * from customer")
			// if err != nil {
			// 	panic(err)
			// }
			// defer rows.Close()
			// Customer_struct_s := []Customer_struct{}

			// for rows.Next() {
			// 	p := Customer_struct{}
			// 	err := rows.Scan(&p.Customer_id, &p.Customer_name, &p.Customer_type, &p.Customer_email)
			// 	if err != nil {
			// 		fmt.Println(err)
			// 		continue
			// 	}
			// 	Customer_struct_s = append(Customer_struct_s, p)
			// }
			// for _, p := range Customer_struct_s {
			// 	customer_map_s[p.Customer_id] = p
			// }

			tmpl.ExecuteTemplate(w, "settings_email", data)

		} else {
			tmpl.ExecuteTemplate(w, "settings_email", data)
		}

		// data := ViewData{
		// 	Title:     "list customer",
		// 	Message:   "list customer below",
		// 	Customers: customer_map,
		// }

		//tmpl.ExecuteTemplate(w, "list_customer", data)

		// Add fill elements form from a global variable or database
		// Add the ability to select an smtp-server or extract a server from an email address
		//http.ServeFile(w, r, "./mail_smtp/settings.html")
	} else {
		email := r.FormValue("email")
		password := r.FormValue("password")

		//fmt.Fprint(w, email+"error auth user not find"+password)

		mass_settings[0] = email
		mass_settings[1] = password

		http.Redirect(w, r, "/", 302)
	}
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
		ErrorLogger.Println(err.Error())
		fmt.Fprint(w, "error"+err.Error())
	} else {
		http.Redirect(w, r, "/", 302)
	}

}

func EditPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	Customer_struct_out := Customer_struct{}
	if type_memory_storage == "SQLit" {
		row := database.QueryRow("select * from customer where customer_id = ?", id)

		err := row.Scan(&Customer_struct_out.Customer_id, &Customer_struct_out.Customer_name, &Customer_struct_out.Customer_type, &Customer_struct_out.Customer_email)
		if err != nil {
			ErrorLogger.Println(err.Error())
			http.Error(w, http.StatusText(404), http.StatusNotFound)
		}

	} else {
		Customer_struct_out = customer_map[id]
	}

	tmpl, err := template.ParseFiles("templates/edit.html", "templates/header.html")
	if err != nil {
		ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl.ExecuteTemplate(w, "edit", Customer_struct_out)

}

func EditHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
	}
	customer_id := r.FormValue("customer_id")
	customer_name := r.FormValue("customer_name")
	customer_type := r.FormValue("customer_type")
	customer_email := r.FormValue("customer_email")

	if type_memory_storage == "SQLit" {
		_, err = database.Exec("update customer set customer_name=?, customer_type=?, customer_email=? where customer_id=?",
			customer_name, customer_type, customer_email, customer_id)

		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}
	} else {
		Customer_struct_out := Customer_struct{}
		Customer_struct_out.Customer_id = customer_id
		Customer_struct_out.Customer_name = customer_name
		Customer_struct_out.Customer_type = customer_type
		Customer_struct_out.Customer_email = customer_email

		customer_map[customer_id] = Customer_struct_out
	}
	http.Redirect(w, r, "/list_customer", 301)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if type_memory_storage == "SQLit" {
		_, err := database.Exec("delete from customer where customer_id = ?", id)
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}
	} else {
		_, ok := customer_map[id]
		if ok {
			delete(customer_map, id)
		}
	}

	http.Redirect(w, r, "/list_customer", 301)
}

func checkINN(w http.ResponseWriter, r *http.Request) {

	customer_INN := r.URL.Query().Get("customer_INN")

	// ////Work with SOAP "github.com/tiaguinho/gosoap"

	client := &http.Client{}

	//replace string
	soapQuery := string(`<?xml version="1.0" encoding="UTF-8"?>
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:req="http://ws.unisoft/FNSNDSCAWS2/Request">
	   <soapenv:Header/>
	   <soapenv:Body>
		  <req:NdsRequest2>
			 <!--1 to 10000 repetitions:-->
			 <req:NP INN="customer_INN"/>
		  </req:NdsRequest2>
	   </soapenv:Body>
	</soapenv:Envelope>`)

	soapQuery = strings.Replace(soapQuery, "customer_INN", customer_INN, 1)

	urlReq := "https://npchk.nalog.ru:443/FNSNDSCAWS_2"

	req, err := http.NewRequest("POST", urlReq, bytes.NewBuffer([]byte(soapQuery)))
	if err != nil {
		ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
	}

	req.ContentLength = int64(len(soapQuery))

	req.Header.Add("Content-Type", "text/xml;charset=UTF-8")
	req.Header.Add("Accept", "text/xml")
	req.Header.Add("SOAPAction", "NdsRequest2")

	resp, err := client.Do(req)
	if err != nil {
		ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
	}

	// don't understand how to create valid xml type map from an body xml
	// res := &Envelope{}
	// err = xml.Unmarshal(body, res)
	// if err != nil {
	// 	fmt.Fprintf(w, err.Error())
	// }

	fmt.Println(string(body))

	result_check := ""

	re := regexp.MustCompile(`State=["]([^"]+)["]`)
	submatchall := re.FindAllStringSubmatch(string(body), -1)
	for _, element := range submatchall {
		result_check = element[1]
	}

	switch result_check {
	case "0":
		fmt.Fprintf(w, "Налогоплательщик зарегистрирован в ЕГРН и имел статус действующего в указанную дату")
	case "1":
		fmt.Fprintf(w, "Налогоплательщик зарегистрирован в ЕГРН, но не имел статус действующего в указанную дату")
	case "2":
		fmt.Fprintf(w, "Налогоплательщик зарегистрирован в ЕГРН")
	case "3":
		fmt.Fprintf(w, "Налогоплательщик с указанным ИНН зарегистрирован в ЕГРН, КПП не соответствует ИНН или не указан*")
	case "4":
		fmt.Fprintf(w, "Налогоплательщик с указанным ИНН не зарегистрирован в ЕГРН")
	case "5":
		fmt.Fprintf(w, "Некорректный ИНН")
	case "6":
		fmt.Fprintf(w, "Недопустимое количество символов ИНН")
	case "7":
		fmt.Fprintf(w, "Недопустимое количество символов КПП")
	case "8":
		fmt.Fprintf(w, "Недопустимые символы в ИНН")
	case "9":
		fmt.Fprintf(w, "Недопустимые символы в КПП")
	case "11":
		fmt.Fprintf(w, "некорректный формат даты")
	case "12":
		fmt.Fprintf(w, "некорректная дата (ранее 01.01.1991 или позднее текущей даты)")
	default:
		fmt.Fprintf(w, "Error find: "+result_check)
	}

}

func initDB() {

	// CREATE TABLE "customer" (
	// 	"customer_id"	TEXT NOT NULL,
	// 	"customer_name"	TEXT,
	// 	"customer_type"	TEXT,
	// 	"customer_email"	TEXT,
	// 	PRIMARY KEY("customer_id")
	// );
	sql_query := "create table customer (customer_id text primary key, customer_name text, customer_type text, customer_email text);"
	_, err := database.Exec(sql_query)
	if err != nil {
		ErrorLogger.Println(err.Error())
		fmt.Println("can't create table : " + err.Error())
	}

	// CREATE TABLE "cookie" (
	// 	"id"	TEXT NOT NULL,
	// 	"user"	TEXT,
	// 	PRIMARY KEY("id")
	// );
	sql_query = "create table cookie (id text primary key, user text);"
	_, err = database.Exec(sql_query)
	if err != nil {
		ErrorLogger.Println(err.Error())
		fmt.Println("can't create table : " + err.Error())
	}

	// CREATE TABLE "users" (
	// 	"user"	TEXT NOT NULL,
	// 	"password"	TEXT,
	// 	PRIMARY KEY("user")
	// );
	sql_query = "create table users (user text primary key, password text);"
	_, err = database.Exec(sql_query)
	if err != nil {
		ErrorLogger.Println(err.Error())
		fmt.Println("can't create table : " + err.Error())
	}

}

func connection_rest_1C(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		// // get parametrs from get-http
		// for key, value := range r.Header {
		// 	if key == "Token" {
		// 		fmt.Println("Token:" + value[0])
		// 	}
		// }

		if type_memory_storage == "SQLit" {

			var customer_map_s = make(map[string]Customer_struct)

			rows, err := database.Query("select * from customer")
			if err != nil {
				ErrorLogger.Println(err.Error())
				panic(err)
			}
			defer rows.Close()
			Customer_struct_s := []Customer_struct{}

			for rows.Next() {
				p := Customer_struct{}
				err := rows.Scan(&p.Customer_id, &p.Customer_name, &p.Customer_type, &p.Customer_email)
				if err != nil {
					ErrorLogger.Println(err.Error())
					fmt.Println(err)
					continue
				}
				Customer_struct_s = append(Customer_struct_s, p)
			}
			for _, p := range Customer_struct_s {
				customer_map_s[p.Customer_id] = p
			}

			userVar2, err := json.Marshal(customer_map_s)
			if err != nil {
				ErrorLogger.Println(err.Error())
				fmt.Fprintf(w, "error json:"+err.Error())
			}
			fmt.Fprintf(w, string(userVar2))

		} else {
			userVar2, err := json.Marshal(customer_map)
			if err != nil {
				ErrorLogger.Println(err.Error())
				fmt.Fprintf(w, "error json:"+err.Error())
			}
			fmt.Fprintf(w, string(userVar2))
		}

	} else {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		var customer_map_json = make(map[string]Customer_struct)
		//Customer_struct := &Customer_struct{}

		err = json.Unmarshal(body, &customer_map_json)
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		if type_memory_storage == "SQLit" {

			var count int

			for _, p := range customer_map_json {

				row := database.QueryRow("select COUNT(*) from customer where customer_id = ?", p.Customer_id)

				err := row.Scan(&count)
				if err != nil {
					ErrorLogger.Println(err.Error())
					fmt.Fprintf(w, err.Error())
				}

				if count == 0 {

					_, err = database.Exec("insert into customer (customer_id, customer_name, customer_type, customer_email) values (?, ?, ?, ?)",
						p.Customer_id, p.Customer_name, p.Customer_type, p.Customer_email)

					if err != nil {
						ErrorLogger.Println(err.Error())
						fmt.Fprintf(w, err.Error())
					}
				} else {
					_, err = database.Exec("update customer set customer_name=?, customer_type=?, customer_email=? where customer_id=?",
						p.Customer_name, p.Customer_type, p.Customer_email, p.Customer_id)

					if err != nil {
						ErrorLogger.Println(err.Error())
						fmt.Fprintf(w, err.Error())
					}
				}
			}

		} else {
			for _, p := range customer_map_json {
				customer_map[p.Customer_id] = p
			}
		}

		fmt.Fprintf(w, string(body))

	}

}

func initLog() {
	file, err := os.OpenFile("./logs/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		ErrorLogger.Println(err.Error())
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger.Println("Starting the application...")
}

func main() {

	initLog()

	type_memory_storage_flag := flag.String("type_memory_storage", "", "type storage data")
	flag.Parse()

	if *type_memory_storage_flag == "" {
		type_memory_storage = "global_variable"
	} else {
		type_memory_storage = *type_memory_storage_flag
	}

	//temporary
	type_memory_storage = "SQLit"

	if type_memory_storage == "SQLit" {

		db, err := sql.Open("sqlite3", "./bd/SQLit/base_sqlit.db")

		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		database = db

		initDB()
		defer db.Close()
	} else {
		users["admin"] = "admin"
	}

	router := mux.NewRouter()

	router.HandleFunc("/", indexHandler)

	//localhost:8181/user?name=Sam&age=21
	router.HandleFunc("/user", user)

	//localhost:8181/get_customer?customer_id="123"
	router.HandleFunc("/get_customer", get_customer)

	//http://localhost:8181/checkINN?customer_INN=7702807750
	router.HandleFunc("/checkINN", checkINN)

	// router.HandleFunc("/users", users)

	router.HandleFunc("/add_change_customer", add_change_customer)
	router.HandleFunc("/postform_add_change_customer", postform_add_change_customer)

	router.HandleFunc("/list_customer", list_customer)

	router.HandleFunc("/mainpage", mainpage)

	router.HandleFunc("/login", login)
	router.HandleFunc("/loginPost", loginPost)

	router.HandleFunc("/email_settings", email_settings)
	//router.HandleFunc("/email_settingsPost", email_settingsPost)

	router.HandleFunc("/send_message", send_message)

	//localhost:8181/edit/2
	router.HandleFunc("/edit/{id:[0-9]+}", EditPage).Methods("GET")
	router.HandleFunc("/edit/{id:[0-9]+}", EditHandler).Methods("POST")
	router.HandleFunc("/delete/{id:[0-9]+}", DeleteHandler)

	router.HandleFunc("/connection_rest_1C", connection_rest_1C)

	// var dir string
	// flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	// flag.Parse()

	//router.Handle("/js/", http.FileServer(http.Dir("./js/")))
	//Работает
	router.PathPrefix("/js").Handler(http.StripPrefix("/js", http.FileServer(http.Dir("./js/"))))

	http.Handle("/", router)

	fmt.Println("Server is listening777...")
	http.ListenAndServe(":8181", nil)

}
