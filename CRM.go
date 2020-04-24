package main

import (
	"crypto/rand"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"log"
	"os"

	"github.com/go-redis/redis/v7"
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

	"context"

	DBLocal "./bd" //add extermal go module.
	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Customer_struct struct {
	Customer_id    string
	Customer_name  string
	Customer_type  string
	Customer_email string
}

// not use
type Customer_struct_bson struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Customer_id    string             `bson:"Customer_id,omitempty"`
	Customer_name  string             `bson:"Customer_name,omitempty"`
	Customer_type  string             `bson:"Customer_type,omitempty"`
	Customer_email string             `bson:"Customer_email,omitempty"`
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

type CustomerStruct_xmlRoot struct {
	Customer_struct []CustomerStruct_xml `xml:"Customer_struct"`
}

type CustomerStruct_xml struct {
	CustomerID    string `xml:"Customer_id"`
	CustomerName  string `xml:"Customer_name"`
	CustomerType  string `xml:"Customer_type"`
	CustomerEmail string `xml:"Customer_email"`
}

var database *sql.DB

var collectionMongoDB *mongo.Collection

var RedisClient *redis.Client

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

func GetCollectionMongoBD(Database string, Collection string, HostConnect string) *mongo.Collection {

	clientOptions := options.Client().ApplyURI(HostConnect)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		ErrorLogger.Println(err.Error())
	}
	err = client.Connect(context.Background())
	if err != nil {
		ErrorLogger.Println(err.Error())
	}

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		ErrorLogger.Println("Couldn't connect to the database", err.Error())
	} else {
		InfoLogger.Println("Connected MongoDB!")
	}

	return client.Database(Database).Collection(Collection)
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

	switch type_memory_storage {
	case "SQLit":

		_, err := database.Exec("insert into customer (customer_id, customer_name, customer_type, customer_email) values (?, ?, ?, ?)",
			customer_data.Customer_id, customer_data.Customer_name, customer_data.Customer_type, customer_data.Customer_email)

		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}
		http.Redirect(w, r, "list_customer", 301)

	case "MongoDB":

		// // tutorial https://www.mongodb.com/blog/post/quick-start-golang--mongodb--modeling-documents-with-go-data-structures
		// // insert one document
		// customer_data := Customer_struct{
		// 	Customer_name:  "customer_name_test",
		// 	Customer_id:    "customer_id_test",
		// 	Customer_type:  "customer_type_test",
		// 	Customer_email: "customer_email_test",
		// }

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		//maybe use? insertMany(): добавляет несколько документов
		//before adding find db.users.find()
		insertResult, err := collectionMongoDB.InsertOne(ctx, customer_data)
		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		fmt.Println(insertResult.InsertedID)

	default:
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

	switch type_memory_storage {
	case "SQLit":

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

	case "MongoDB":

		var customer_map_s = make(map[string]Customer_struct)

		cur, err := collectionMongoDB.Find(context.Background(), bson.D{})
		if err != nil {
			ErrorLogger.Println(err.Error())
		}
		defer cur.Close(context.Background())

		Customer_struct_slice := []Customer_struct{}

		for cur.Next(context.Background()) {

			Customer_struct_out := Customer_struct{}

			err := cur.Decode(&Customer_struct_out)
			if err != nil {
				ErrorLogger.Println(err.Error())
			}

			Customer_struct_slice = append(Customer_struct_slice, Customer_struct_out)

			// To get the raw bson bytes use cursor.Current
			// // raw := cur.Current
			// // fmt.Println(raw)
			// do something with raw...
		}
		if err := cur.Err(); err != nil {
			ErrorLogger.Println(err.Error())
		}

		for _, p := range Customer_struct_slice {
			customer_map_s[p.Customer_id] = p
		}

		tmpl.ExecuteTemplate(w, "list_customer", customer_map_s)

	default:
		tmpl.ExecuteTemplate(w, "list_customer", customer_map)
	}

}

func mainpage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Index Page")
}

func RedirectToHTTPS(w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, "https://localhost:8182"+r.RequestURI,
		http.StatusMovedPermanently)

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
	//http.Redirect(w, r, "/", 302)
	http.Redirect(w, r, "http://localhost:8181/",
		http.StatusMovedPermanently)
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
	switch type_memory_storage {
	case "SQLit":

		row := database.QueryRow("select * from customer where customer_id = ?", id)

		err := row.Scan(&Customer_struct_out.Customer_id, &Customer_struct_out.Customer_name, &Customer_struct_out.Customer_type, &Customer_struct_out.Customer_email)
		if err != nil {
			ErrorLogger.Println(err.Error())
			http.Error(w, http.StatusText(404), http.StatusNotFound)
		}

	case "MongoDB":

		err := collectionMongoDB.FindOne(context.TODO(), bson.D{{"customer_id", id}}).Decode(&Customer_struct_out)
		if err != nil {
			// ErrNoDocuments means that the filter did not match any documents in the collection
			if err == mongo.ErrNoDocuments {
				return
			}
			log.Fatal(err)
		}
		fmt.Printf("found document %v", Customer_struct_out)

	default:
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

	switch type_memory_storage {
	case "SQLit":

		_, err = database.Exec("update customer set customer_name=?, customer_type=?, customer_email=? where customer_id=?",
			customer_name, customer_type, customer_email, customer_id)

		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

	case "MongoDB":

		opts := options.Update().SetUpsert(true)
		filter := bson.D{{"customer_id", customer_id}}
		update := bson.D{{"$set", bson.D{{"customer_name", customer_name}, {"customer_type", customer_type}, {"customer_email", customer_email}}}}

		result, err := collectionMongoDB.UpdateOne(context.TODO(), filter, update, opts)
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		if result.MatchedCount != 0 {
			fmt.Println("matched and replaced an existing document")
		}
		if result.UpsertedCount != 0 {
			fmt.Printf("inserted a new document with ID %v\n", result.UpsertedID)
		}

	default:
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

	switch type_memory_storage {
	case "SQLit":
		_, err := database.Exec("delete from customer where customer_id = ?", id)
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}
	case "MongoDB":

		res, err := collectionMongoDB.DeleteOne(context.TODO(), bson.D{{"customer_id", id}})
		if err != nil {
			ErrorLogger.Println(err.Error())
		}
		fmt.Printf("deleted %v documents\n", res.DeletedCount)

	default:
		_, ok := customer_map[id]
		if ok {
			delete(customer_map, id)
		}
	}

	http.Redirect(w, r, "/list_customer", 301)

}

func checkINN(w http.ResponseWriter, r *http.Request) {

	customer_INN := r.URL.Query().Get("customer_INN")

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

	// maybe consider opportunity using the package  https://github.com/beevik/etree
	// to build and parse xml for SOAP
	// below example
	// doc := etree.NewDocument()
	// if err := doc.ReadFromString(soapQuery); err != nil {
	// 	panic(err)
	// }

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

func initDBSQLit() {

	// CREATE TABLE "customer" (
	// 	"customer_id"	TEXT NOT NULL,
	// 	"customer_name"	TEXT,
	// 	"customer_type"	TEXT,
	// 	"customer_email"	TEXT,
	// 	PRIMARY KEY("customer_id")
	// );
	sql_query := "create table if not exists customer (customer_id text primary key, customer_name text, customer_type text, customer_email text);"
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
	sql_query = "create table if not exists cookie (id text primary key, user text);"
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
	sql_query = "create table if not exists users (user text primary key, password text);"
	_, err = database.Exec(sql_query)
	if err != nil {
		ErrorLogger.Println(err.Error())
		fmt.Println("can't create table : " + err.Error())
	}

}

func Api_json(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		// // get parametrs from get-http
		// for key, value := range r.Header {
		// 	if key == "Token" {
		// 		fmt.Println("Token:" + value[0])
		// 	}
		// }

		switch type_memory_storage {
		case "SQLit":

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

		case "MongoDB":

			var customer_map_s = make(map[string]Customer_struct)

			cur, err := collectionMongoDB.Find(context.Background(), bson.D{})
			if err != nil {
				ErrorLogger.Println(err.Error())
			}
			defer cur.Close(context.Background())

			Customer_struct_slice := []Customer_struct{}

			for cur.Next(context.Background()) {

				Customer_struct_out := Customer_struct{}

				err := cur.Decode(&Customer_struct_out)
				if err != nil {
					ErrorLogger.Println(err.Error())
				}

				Customer_struct_slice = append(Customer_struct_slice, Customer_struct_out)

				// To get the raw bson bytes use cursor.Current
				// // raw := cur.Current
				// // fmt.Println(raw)
				// do something with raw...
			}
			if err := cur.Err(); err != nil {
				ErrorLogger.Println(err.Error())
			}

			for _, p := range Customer_struct_slice {
				customer_map_s[p.Customer_id] = p
			}

			userVar2, err := json.Marshal(customer_map_s)
			if err != nil {
				ErrorLogger.Println(err.Error())
				fmt.Fprintf(w, "error json:"+err.Error())
			}
			fmt.Fprintf(w, string(userVar2))

		case "Redis":

			var customer_map_s = make(map[string]Customer_struct)
			Customer_struct_slice := []Customer_struct{}

			// find a function that gets all the keys to Reddit
			i := 0
			for {
				p := Customer_struct{}
				IDString := strconv.FormatInt(int64(i), 10)
				val2, err := RedisClient.Get(IDString).Result()
				if err == redis.Nil {
					//fmt.Println("key2 does not exist")
				} else if err != nil {
					panic(err)
				} else {
					fmt.Println("key2", val2)

					err = json.Unmarshal([]byte(val2), &p)
					if err != nil {
						ErrorLogger.Println(err.Error())
						fmt.Fprintf(w, err.Error())
					}

					Customer_struct_slice = append(Customer_struct_slice, p)
				}
				i++
				if i > 1000 {
					break
				}
			}

			for _, p := range Customer_struct_slice {
				customer_map_s[p.Customer_id] = p
			}

			userVar2, err := json.Marshal(customer_map_s)
			if err != nil {
				ErrorLogger.Println(err.Error())
				fmt.Fprintf(w, "error json:"+err.Error())
			}
			fmt.Fprintf(w, string(userVar2))

		default:
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

		switch type_memory_storage {
		case "SQLit":

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

		case "MongoDB":

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

			//maybe use? insertMany(): добавляет несколько документов
			//before adding find db.users.find()

			for _, p := range customer_map_json {
				insertResult, err := collectionMongoDB.InsertOne(ctx, p)
				if err != nil {
					ErrorLogger.Println(err.Error())
					panic(err)
				}
				fmt.Println(insertResult.InsertedID)
			}

		case "Redis":

			for _, p := range customer_map_json {

				userVar2, err := json.Marshal(p)
				if err != nil {
					ErrorLogger.Println(err.Error())
					fmt.Fprintf(w, "error json:"+err.Error())
				}

				err = RedisClient.Set(p.Customer_id, string(userVar2), 0).Err()
				if err != nil {
					panic(err)
				}

			}

		default:
			for _, p := range customer_map_json {
				customer_map[p.Customer_id] = p
			}
		}

		fmt.Fprintf(w, string(body))

	}

}

func Api_xml(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		switch type_memory_storage {
		case "SQLit":

			// var customer_map_s = make(map[string]Customer_struct)

			// rows, err := database.Query("select * from customer")
			// if err != nil {
			// 	ErrorLogger.Println(err.Error())
			// 	panic(err)
			// }
			// defer rows.Close()
			// Customer_struct_s := []Customer_struct{}

			// for rows.Next() {
			// 	p := Customer_struct{}
			// 	err := rows.Scan(&p.Customer_id, &p.Customer_name, &p.Customer_type, &p.Customer_email)
			// 	if err != nil {
			// 		ErrorLogger.Println(err.Error())
			// 		fmt.Println(err)
			// 		continue
			// 	}
			// 	Customer_struct_s = append(Customer_struct_s, p)
			// }
			// for _, p := range Customer_struct_s {
			// 	customer_map_s[p.Customer_id] = p
			// }

			// userVar2, err := json.Marshal(customer_map_s)
			// if err != nil {
			// 	ErrorLogger.Println(err.Error())
			// 	fmt.Fprintf(w, "error json:"+err.Error())
			// }
			// fmt.Fprintf(w, string(userVar2))

		case "MongoDB":

			var customer_map_s = make(map[string]Customer_struct)

			cur, err := collectionMongoDB.Find(context.Background(), bson.D{})
			if err != nil {
				ErrorLogger.Println(err.Error())
			}
			defer cur.Close(context.Background())

			Customer_struct_slice := []Customer_struct{}

			for cur.Next(context.Background()) {

				Customer_struct_out := Customer_struct{}

				err := cur.Decode(&Customer_struct_out)
				if err != nil {
					ErrorLogger.Println(err.Error())
				}

				Customer_struct_slice = append(Customer_struct_slice, Customer_struct_out)

			}
			if err := cur.Err(); err != nil {
				ErrorLogger.Println(err.Error())
			}

			for _, p := range Customer_struct_slice {
				customer_map_s[p.Customer_id] = p
			}

			// userVar2, err := json.Marshal(customer_map_s)
			// if err != nil {
			// 	ErrorLogger.Println(err.Error())
			// 	fmt.Fprintf(w, "error json:"+err.Error())
			// }
			// fmt.Fprintf(w, string(userVar2))

			xmlData, _ := xml.MarshalIndent(Customer_struct_slice, " ", "  ")
			fmt.Fprintf(w, string(xmlData))

			test_rez_slice := []CustomerStruct_xml{}
			//var test_rez []Customer_struct
			if err := xml.Unmarshal(xmlData, &test_rez_slice); err != nil {
				panic(err)
			}
			fmt.Println(test_rez_slice)

		default:
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

		err = json.Unmarshal(body, &customer_map_json)
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		switch type_memory_storage {
		case "SQLit":

			// var count int

			// for _, p := range customer_map_json {

			// 	row := database.QueryRow("select COUNT(*) from customer where customer_id = ?", p.Customer_id)

			// 	err := row.Scan(&count)
			// 	if err != nil {
			// 		ErrorLogger.Println(err.Error())
			// 		fmt.Fprintf(w, err.Error())
			// 	}

			// 	if count == 0 {

			// 		_, err = database.Exec("insert into customer (customer_id, customer_name, customer_type, customer_email) values (?, ?, ?, ?)",
			// 			p.Customer_id, p.Customer_name, p.Customer_type, p.Customer_email)

			// 		if err != nil {
			// 			ErrorLogger.Println(err.Error())
			// 			fmt.Fprintf(w, err.Error())
			// 		}
			// 	} else {
			// 		_, err = database.Exec("update customer set customer_name=?, customer_type=?, customer_email=? where customer_id=?",
			// 			p.Customer_name, p.Customer_type, p.Customer_email, p.Customer_id)

			// 		if err != nil {
			// 			ErrorLogger.Println(err.Error())
			// 			fmt.Fprintf(w, err.Error())
			// 		}
			// 	}
			// }

		case "MongoDB":

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

			//maybe use? insertMany(): добавляет несколько документов
			//before adding find db.users.find()

			for _, p := range customer_map_json {
				insertResult, err := collectionMongoDB.InsertOne(ctx, p)
				if err != nil {
					ErrorLogger.Println(err.Error())
					panic(err)
				}
				fmt.Println(insertResult.InsertedID)
			}

			//	//	res := &Result{}
			//	//string(contents_2)
			//	//err2 := xml.Unmarshal(Res_test.Bytes(), res)
			//	//	err2 := xml.Unmarshal(contents, res)
			//	//	if err2 != nil {
			//	//		fmt.Println("Error Unmarshal = ", err2.Error())
			//	//	}

		default:
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

func infoTeam(rw http.ResponseWriter, r *http.Request) {

	fmt.Println("Main Page")
	//rw.Header().Set(" Server in Golang")
	rw.Write([]byte("Service (Golang Server)\nTeam Members:\n Agha Assad\n"))
}

func intiRedisClient(Addr string) *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return client
}

func main() {

	//fmt.Println(DBLocal.Test(5))

	initLog()

	type_memory_storage_flag := flag.String("type_memory_storage", "", "type storage data")
	flag.Parse()

	if *type_memory_storage_flag == "" {
		//type_memory_storage = "global_variable"

		//temporary
		type_memory_storage = "Redis"
	} else {
		type_memory_storage = *type_memory_storage_flag
	}

	switch type_memory_storage {
	case "SQLit":
		db, err := sql.Open("sqlite3", "./bd/SQLit/base_sqlit.db")

		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		database = db

		initDBSQLit()
		defer db.Close()

	case "Redis":

		RedisClient = intiRedisClient("localhost:32773")

		pong, err := RedisClient.Ping().Result()
		if err != nil {
			panic(err)
		}
		fmt.Println(pong, err)

	case "MongoDB":

		//temporary
		//collectionMongoDB = GetCollectionMongoBD("CRM", "customers", "mongodb://localhost:32768")
		collectionMongoDB = DBLocal.GetCollectionMongoBD("CRM", "customers", "mongodb://localhost:32768")

	default:
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

	// replace to HTTPS router
	router.HandleFunc("/login", RedirectToHTTPS)
	router.HandleFunc("/loginPost", RedirectToHTTPS)

	router.HandleFunc("/email_settings", email_settings)
	//router.HandleFunc("/email_settingsPost", email_settingsPost)

	router.HandleFunc("/send_message", send_message)

	//localhost:8181/edit/2
	router.HandleFunc("/edit/{id:[0-9]+}", EditPage).Methods("GET")
	router.HandleFunc("/edit/{id:[0-9]+}", EditHandler).Methods("POST")
	router.HandleFunc("/delete/{id:[0-9]+}", DeleteHandler)

	router.HandleFunc("/api_json", Api_json)
	router.HandleFunc("/api_xml", Api_xml)

	// var dir string
	// flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	// flag.Parse()

	//router.Handle("/js/", http.FileServer(http.Dir("./js/")))
	//Работает
	router.PathPrefix("/js").Handler(http.StripPrefix("/js", http.FileServer(http.Dir("./js/"))))

	router_HTTPS := mux.NewRouter()
	router_HTTPS.HandleFunc("/login", login)
	router_HTTPS.HandleFunc("/loginPost", loginPost)

	httpsMux := http.NewServeMux()
	//httpsMux.Handle("/", http.HandlerFunc(infoTeam))
	httpsMux.Handle("/", router_HTTPS)
	go http.ListenAndServeTLS(":8182", "./Cert/cert.pem", "./Cert/key.pem", httpsMux)

	http.Handle("/", router)
	http.ListenAndServe(":8181", nil)
	fmt.Println("Server is listening777...")

}
