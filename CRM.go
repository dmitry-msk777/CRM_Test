package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"log"
	"os"

	"github.com/beevik/etree"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"

	//"github.com/gorilla/sessions"
	"encoding/base64"
	"net/mail"
	"net/smtp"

	//"github.com/tiaguinho/gosoap"
	"bytes"
	"encoding/json"
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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/olivere/elastic"

	"net"

	pb "../CRM_Test/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type EngineCRM struct {
	DataBaseType      string
	collectionMongoDB *mongo.Collection
	DemoDBmap         map[string]Customer_struct
	databaseSQLite    *sql.DB
	RedisClient       *redis.Client
	Global_settings   Global_settings
	Users_CRM_map     map[string]Users_CRM
	Cookie_CRM_map    map[string]Cookie_CRM
}

func (EngineCRM *EngineCRM) SetSettings(Global_settings Global_settings) {

	EngineCRM.DataBaseType = Global_settings.DataBaseType
	if Global_settings.DataBaseType == "" {
		EngineCRM.DataBaseType = "DemoRegime"
		Global_settingsV.DataBaseType = "DemoRegime"
	}

	EngineCRM.Global_settings = Global_settingsV

}

func (EngineCRM *EngineCRM) GetOneJSON() string {
	JsonString, err := json.Marshal(EngineCRM.DemoDBmap)
	if err != nil {
		return err.Error()
	}

	return string(JsonString)
}

func (EngineCRM *EngineCRM) InitDataBase() bool {

	switch EngineCRMv.DataBaseType {
	case "SQLit":
		db, err := sql.Open("sqlite3", "./bd/SQLit/base_sqlit.db")

		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		database = db
		EngineCRMv.databaseSQLite = db

		// CREATE TABLE "customer" (
		// 	"customer_id"	TEXT NOT NULL,
		// 	"customer_name"	TEXT,
		// 	"customer_type"	TEXT,
		// 	"customer_email"	TEXT,
		// 	PRIMARY KEY("customer_id")
		// );
		sql_query := "create table if not exists customer (customer_id text primary key, customer_name text, customer_type text, customer_email text);"
		_, err = EngineCRMv.databaseSQLite.Exec(sql_query)
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
		_, err = EngineCRMv.databaseSQLite.Exec(sql_query)
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
		_, err = EngineCRMv.databaseSQLite.Exec(sql_query)
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Println("can't create table : " + err.Error())
		}

	case "Redis":
		//localhost:32769
		EngineCRMv.RedisClient = intiRedisClient(EngineCRMv.Global_settings.AddressRedis)

		pong, err := EngineCRMv.RedisClient.Ping().Result()
		if err != nil {
			EngineCRMv.RedisClient = nil
			fmt.Println(pong, err)
			return false
		}

	case "MongoDB":

		//temporary
		//collectionMongoDB = GetCollectionMongoBD("CRM", "customers", "mongodb://localhost:32768")
		//"mongodb://localhost:32768"
		EngineCRMv.collectionMongoDB = DBLocal.GetCollectionMongoBD("CRM", "customers", EngineCRMv.Global_settings.AddressMongoBD)

	default:

		var ArrayCustomer []Customer_struct

		ArrayCustomer = append(ArrayCustomer, Customer_struct{
			Customer_id:    "777",
			Customer_name:  "Dmitry",
			Customer_type:  "Cust",
			Customer_email: "fff@mail.ru",
		})

		ArrayCustomer = append(ArrayCustomer, Customer_struct{
			Customer_id:    "666",
			Customer_name:  "Alex",
			Customer_type:  "Cust_Fiz",
			Customer_email: "44fish@mail.ru",
		})

		var mapForEngineCRM = make(map[string]Customer_struct)
		EngineCRM.DemoDBmap = mapForEngineCRM

		var users_CRM_def = make(map[string]Users_CRM)
		EngineCRM.Users_CRM_map = users_CRM_def

		users_CRM_data := Users_CRM{
			user:     "admin",
			password: "1234"}

		EngineCRM.Users_CRM_map["admin"] = users_CRM_data

		var cookie_CRM_def = make(map[string]Cookie_CRM)
		EngineCRM.Cookie_CRM_map = cookie_CRM_def

		for _, p := range ArrayCustomer {
			EngineCRM.DemoDBmap[p.Customer_id] = p
		}

	}

	return true
}

func (EngineCRM *EngineCRM) GetAllCustomer(DataBaseType string) map[string]Customer_struct {

	var customer_map_s = make(map[string]Customer_struct)

	switch DataBaseType {
	case "SQLit":

		rows, err := EngineCRM.databaseSQLite.Query("select * from customer")
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

		return customer_map_s

	case "MongoDB":

		cur, err := EngineCRMv.collectionMongoDB.Find(context.Background(), bson.D{})
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

		return customer_map_s

	case "Redis":

		Customer_struct_slice := []Customer_struct{}

		// find a function that gets all the keys to Reddit
		i := 0
		for {
			p := Customer_struct{}
			IDString := strconv.FormatInt(int64(i), 10)
			val2, err := EngineCRMv.RedisClient.Get(IDString).Result()
			if err == redis.Nil {
				//fmt.Println("key2 does not exist")
			} else if err != nil {
				panic(err)
			} else {
				fmt.Println("key2", val2)

				err = json.Unmarshal([]byte(val2), &p)
				if err != nil {
					ErrorLogger.Println(err.Error())
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

		return customer_map_s

	default:
		return EngineCRM.DemoDBmap
	}

}

func (EngineCRM *EngineCRM) FindOneRow(DataBaseType string, id string) Customer_struct {

	Customer_struct_out := Customer_struct{}

	switch DataBaseType {
	case "SQLit":

		row := EngineCRMv.databaseSQLite.QueryRow("select * from customer where customer_id = ?", id)

		err := row.Scan(&Customer_struct_out.Customer_id, &Customer_struct_out.Customer_name, &Customer_struct_out.Customer_type, &Customer_struct_out.Customer_email)
		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}

	case "MongoDB":

		err := EngineCRMv.collectionMongoDB.FindOne(context.TODO(), bson.D{{"customer_id", id}}).Decode(&Customer_struct_out)
		if err != nil {
			// ErrNoDocuments means that the filter did not match any documents in the collection
			if err == mongo.ErrNoDocuments {
				panic(err)
			}
			log.Fatal(err)
		}
		fmt.Printf("found document %v", Customer_struct_out)

	case "Redit":

	default:
		Customer_struct_out = EngineCRMv.DemoDBmap[id]
	}

	return Customer_struct_out
}

func (EngineCRM *EngineCRM) AddChangeOneRow(DataBaseType string, Customer_struct Customer_struct) string {

	switch DataBaseType {
	case "SQLit":

		var count int

		row := EngineCRMv.databaseSQLite.QueryRow("select COUNT(*) from customer where customer_id = ?", Customer_struct.Customer_id)

		err := row.Scan(&count)
		if err != nil {
			ErrorLogger.Println(err.Error())
			return err.Error()
		}

		if count == 0 {

			_, err = EngineCRMv.databaseSQLite.Exec("insert into customer (customer_id, customer_name, customer_type, customer_email) values (?, ?, ?, ?)",
				Customer_struct.Customer_id, Customer_struct.Customer_name, Customer_struct.Customer_type, Customer_struct.Customer_email)

			if err != nil {
				ErrorLogger.Println(err.Error())
				return err.Error()
			}
		} else {
			_, err = EngineCRMv.databaseSQLite.Exec("update customer set customer_name=?, customer_type=?, customer_email=? where customer_id=?",
				Customer_struct.Customer_name, Customer_struct.Customer_type, Customer_struct.Customer_email, Customer_struct.Customer_id)

			if err != nil {
				ErrorLogger.Println(err.Error())
				return err.Error()
			}
		}

	case "MongoDB":

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		//maybe use? insertMany(): добавляет несколько документов
		//before adding find db.users.find()

		insertResult, err := EngineCRMv.collectionMongoDB.InsertOne(ctx, Customer_struct)
		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		fmt.Println(insertResult.InsertedID)

		// This update function can use the separate Update and Paste pre-search?

		// opts := options.Update().SetUpsert(true)
		// filter := bson.D{{"customer_id", Customer_struct.Customer_id}}
		// update := bson.D{{"$set", bson.D{{"customer_name", Customer_struct.Customer_name}, {"customer_type", Customer_struct.Customer_type}, {"customer_email", Customer_struct.Customer_email}}}}

		// result, err := EngineCRMv.collectionMongoDB.UpdateOne(context.TODO(), filter, update, opts)
		// if err != nil {
		// 	ErrorLogger.Println(err.Error())
		// 	return err.Error()
		// }

		// if result.MatchedCount != 0 {
		// 	fmt.Println("matched and replaced an existing document")
		// }
		// if result.UpsertedCount != 0 {
		// 	fmt.Printf("inserted a new document with ID %v\n", result.UpsertedID)
		// }
	case "Redis":

		JsonStr, err := json.Marshal(Customer_struct)
		if err != nil {
			ErrorLogger.Println(err.Error())
			return "error json:" + err.Error()
		}

		err = EngineCRMv.RedisClient.Set(Customer_struct.Customer_id, string(JsonStr), 0).Err()
		if err != nil {
			panic(err)
		}

	default:
		EngineCRMv.DemoDBmap[Customer_struct.Customer_id] = Customer_struct
	}

	return ""
}

func (EngineCRM *EngineCRM) DeleteOneRow(DataBaseType string, id string) string {

	switch DataBaseType {
	case "SQLit":
		_, err := EngineCRMv.databaseSQLite.Exec("delete from customer where customer_id = ?", id)
		if err != nil {
			ErrorLogger.Println(err.Error())
			return err.Error()
		}
	case "MongoDB":

		res, err := EngineCRMv.collectionMongoDB.DeleteOne(context.TODO(), bson.D{{"customer_id", id}})
		if err != nil {
			ErrorLogger.Println(err.Error())
		}
		fmt.Printf("deleted %v documents\n", res.DeletedCount)

	case "Redis":

	default:
		_, ok := EngineCRMv.DemoDBmap[id]
		if ok {
			delete(EngineCRMv.DemoDBmap, id)
		}
	}

	return ""

}

var EngineCRMv EngineCRM

type Customer_struct struct {
	Customer_id    string
	Customer_name  string
	Customer_type  string
	Customer_email string
}

type Global_settings struct {
	DataBaseType    string
	Mail_smtpServer string
	AddressMongoBD  string
	AddressRedis    string
	Mail_email      string
	Mail_password   string
}

func (GlobalSettings *Global_settings) SaveSettingsOnDisk() {

	f, err := os.Create("./settings/config.json")
	if err != nil {
		log.Fatal(err)
	}

	JsonString, err := json.Marshal(Global_settingsV)
	if err != nil {
		ErrorLogger.Println(err.Error())
		log.Fatal(err)
	}

	if _, err := f.Write(JsonString); err != nil {
		f.Close() // ignore error; Write error takes precedence
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func (GlobalSettings *Global_settings) LoadSettingsFromDisk() {

	file, err := os.OpenFile("./settings/config.json", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(err)
	}

	decoder := json.NewDecoder(file)
	Settings := Global_settings{}
	err = decoder.Decode(&Settings)
	if err != nil {
		fmt.Println(err)
	}

	Global_settingsV = Settings

	if err := file.Close(); err != nil {
		fmt.Println(err)
	}
}

var Global_settingsV Global_settings

//////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

// not used
type Customer_struct_bson struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Customer_id    string             `bson:"Customer_id,omitempty"`
	Customer_name  string             `bson:"Customer_name,omitempty"`
	Customer_type  string             `bson:"Customer_type,omitempty"`
	Customer_email string             `bson:"Customer_email,omitempty"`
}

type Users_CRM struct {
	user     string
	password string
}

type Cookie_CRM struct {
	id   string
	user string
}

var Cookie_CRMv Cookie_CRM

// convert in cookie_base type
func (Cookie_CRM *Cookie_CRM) GenerateId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

var database *sql.DB
var collectionMongoDB *mongo.Collection
var RedisClient *redis.Client

var CRM_Counter_Prometheus_JSON prometheus.Counter
var CRM_Counter_Prometheus_XML prometheus.Counter
var CRM_Counter_Gauge prometheus.Gauge

var customer_map = make(map[string]Customer_struct)

var cookiemap = make(map[string]string)
var users = make(map[string]string)

var mass_settings = make([]string, 2)

const cookieName = "CookieCRM"

type ViewData struct {
	Title        string
	Message      string
	User         string
	DataBaseType string
	Customers    map[string]Customer_struct
}

var InfoLogger *log.Logger
var ErrorLogger *log.Logger

type server struct{}

func (s *server) GET_List(ctx context.Context, in *pb.RequestGET) (*pb.ResponseGET, error) {

	id := in.CustomerId

	Customer_struct_out := EngineCRMv.FindOneRow(EngineCRMv.DataBaseType, id)

	response := &pb.ResponseGET{
		CustomerId:    Customer_struct_out.Customer_id,
		CustomerName:  Customer_struct_out.Customer_name,
		CustomerType:  Customer_struct_out.Customer_type,
		CustomerEmail: Customer_struct_out.Customer_email,
	}

	return response, nil
}

func (s *server) POST_List(ctx context.Context, in *pb.RequestPOST) (*pb.ResponsePOST, error) {

	Customer_struct_out := Customer_struct{
		Customer_id:    in.CustomerId,
		Customer_name:  in.CustomerName,
		Customer_type:  in.CustomerType,
		Customer_email: in.CustomerEmail,
	}

	EngineCRMv.AddChangeOneRow(EngineCRMv.DataBaseType, Customer_struct_out)

	return &pb.ResponsePOST{CustomerId: "True"}, nil
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
		nameUserFromCookie, flagmap := EngineCRMv.Cookie_CRM_map[CookieGet.Value]
		if flagmap != false {
			nameUserFromCookieStruc = nameUserFromCookie.user
		}
	}

	if EngineCRMv.DataBaseType == "SQLit" && CookieGet != nil {

		rows, err := database.Query("select * from cookie where id = $1", CookieGet.Value)
		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		defer rows.Close()
		cookie_base_s := []Cookie_CRM{}

		for rows.Next() {
			p := Cookie_CRM{}
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

	customer_id_for_find := r.URL.Query().Get("customer_id")

	switch EngineCRMv.DataBaseType {
	case "SQLit":
		fmt.Fprintf(w, "function not implemented for SQLit")
	case "MongoDB":

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

		//ElasticSerch

		clientElasticSerch, err := elastic.NewClient(elastic.SetSniff(false),
			elastic.SetURL("http://127.0.0.1:32771", "http://127.0.0.1:32770"))
		// elastic.SetBasicAuth("user", "secret"))
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		exists, err := clientElasticSerch.IndexExists("crm_customer").Do(context.Background()) //twitter
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		if !exists {
			// Create a new index.
			mapping := `
				{
					"settings":{
						"number_of_shards":1,
						"number_of_replicas":0
					},
					"mappings":{
						"doc":{
							"properties":{
								"Customer_name":{
									"type":"text"
								},
								"Customer_id":{
									"type":"text",
									"store": true,
									"fielddata": true
								},
								"Customer_type":{
									"type":"text"
								},
								"Customer_email":{
									"type":"text"
								}
						}
					}
				}
				}`

			//createIndex, err := clientElasticSerch.CreateIndex("crm_customer").Body(mapping).IncludeTypeName(true).Do(context.Background())
			createIndex, err := clientElasticSerch.CreateIndex("crm_customer").Body(mapping).Do(context.Background())
			if err != nil {
				ErrorLogger.Println(err.Error())
				fmt.Fprintf(w, err.Error())
				return
			}
			if !createIndex.Acknowledged {
			}
		}

		for _, p := range Customer_struct_slice {

			put1, err := clientElasticSerch.Index().
				Index("crm_customer").
				Type("doc").
				Id(p.Customer_id).
				BodyJson(p).
				Do(context.Background())
			if err != nil {
				ErrorLogger.Println(err.Error())
				fmt.Fprintf(w, err.Error())
				return
			}
			fmt.Printf("Indexed tweet %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)

		}

		// Flush to make sure the documents got written.
		_, err = clientElasticSerch.Flush().Index("crm_customer").Do(context.Background())
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		// Search with a term query
		termQuery := elastic.NewTermQuery("Customer_id", customer_id_for_find)
		searchResult, err := clientElasticSerch.Search().
			Index("crm_customer").     // search in index "crm_customer"
			Query(termQuery).          // specify the query
			Sort("Customer_id", true). // sort by "user" field, ascending
			From(0).Size(10).          // take documents 0-9
			Pretty(true).              // pretty print request and response JSON
			Do(context.Background())   // execute
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		// searchResult is of type SearchResult and returns hits, suggestions,
		// and all kinds of other information from Elasticsearch.
		fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

		var ttyp Customer_struct
		for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
			t := item.(Customer_struct)
			fmt.Fprintf(w, "customer_id: %s customer_name: %s", t.Customer_id, t.Customer_name)
		}

		fmt.Printf("Found a total of %d tweets\n", searchResult.TotalHits())

		// Delete an index.
		deleteIndex, err := clientElasticSerch.DeleteIndex("crm_customer").Do(context.Background())
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}
		if !deleteIndex.Acknowledged {
			// Not acknowledged
		}

		//ElasticSerch

	case "Redis":

		fmt.Fprintf(w, "function not implemented for Redis")

	default:
		fmt.Fprintf(w, "customer_id: %s customer_name: %s", customer_map[customer_id_for_find].Customer_id,
			customer_map[customer_id_for_find].Customer_name)
	}

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

	EngineCRMv.AddChangeOneRow(EngineCRMv.DataBaseType, customer_data)

	http.Redirect(w, r, "/list_customer", 302)
}

func list_customer(w http.ResponseWriter, r *http.Request) {

	//prometheus
	CRM_Counter_Gauge.Set(float64(5)) // or: Inc(), Dec(), Add(5), Dec(5),

	tmpl, err := template.ParseFiles("templates/list_customer.html", "templates/header.html")
	if err != nil {
		ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl.ExecuteTemplate(w, "list_customer", EngineCRMv.GetAllCustomer(EngineCRMv.DataBaseType))

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

	if EngineCRMv.DataBaseType == "SQLit" {

		rows, err := EngineCRMv.databaseSQLite.Query("select * from users where user = $1 and password = $2", username, password)
		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		defer rows.Close()
		users_base_s := []Users_CRM{}

		for rows.Next() {
			p := Users_CRM{}
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

		user_password_struct, flagusers := EngineCRMv.Users_CRM_map[username]
		if flagusers == true {
			if user_password_struct.password != password {
				fmt.Fprint(w, "error auth password")
				return
			}
		} else {
			fmt.Fprint(w, "error auth user not find")
			return
		}
	}

	idcookie := Cookie_CRMv.GenerateId()

	if EngineCRMv.DataBaseType == "SQLit" {

		result, err := EngineCRMv.databaseSQLite.Exec("insert into cookie (id, user) values ($1, $2)",
			idcookie, username)
		if err != nil {
			ErrorLogger.Println(err.Error())
			panic(err)
		}
		fmt.Println(result.LastInsertId()) // id последнего добавленного объекта
		fmt.Println(result.RowsAffected()) // количество добавленных строк

	} else {
		cookie_CRM_data := Cookie_CRM{
			id:   idcookie,
			user: username,
		}
		EngineCRMv.Cookie_CRM_map[idcookie] = cookie_CRM_data
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

func settings(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		tmpl, err := template.ParseFiles("templates/settings.html", "templates/header.html")
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		tmpl.ExecuteTemplate(w, "settings", Global_settingsV)

	} else {

		mass_settings[0] = r.FormValue("email")
		mass_settings[1] = r.FormValue("password")

		Global_settingsV.Mail_email = r.FormValue("Mail_email")
		Global_settingsV.Mail_password = r.FormValue("Mail_password")
		Global_settingsV.Mail_smtpServer = r.FormValue("Mail_smtpServer")
		Global_settingsV.DataBaseType = r.FormValue("DataBaseType")

		Global_settingsV.AddressMongoBD = r.FormValue("AddressMongoBD")
		Global_settingsV.AddressRedis = r.FormValue("AddressRedis")

		EngineCRMv.SetSettings(Global_settingsV)

		EngineCRMv.InitDataBase()

		Global_settingsV.SaveSettingsOnDisk()

		http.Redirect(w, r, "/", 302)
	}
}

func send_message(w http.ResponseWriter, r *http.Request) {

	// Set up authentication information. https://yandex.ru/support/mail/mail-clients.html

	//smtpServer := "smtp.yandex.ru"
	smtpServer := Global_settingsV.Mail_smtpServer
	auth := smtp.PlainAuth(
		"",
		Global_settingsV.Mail_email,
		Global_settingsV.Mail_password,
		smtpServer,
	)

	from := mail.Address{"Test", Global_settingsV.Mail_email}
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

	Customer_struct_out := EngineCRMv.FindOneRow(EngineCRMv.DataBaseType, id)

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

	Customer_struct_out := Customer_struct{
		Customer_id:    r.FormValue("customer_id"),
		Customer_name:  r.FormValue("customer_name"),
		Customer_type:  r.FormValue("customer_type"),
		Customer_email: r.FormValue("customer_email"),
	}

	EngineCRMv.AddChangeOneRow(EngineCRMv.DataBaseType, Customer_struct_out)

	//return err
	//fmt.Fprintf(w, err.Error())

	http.Redirect(w, r, "/list_customer", 301)

}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	EngineCRMv.DeleteOneRow(EngineCRMv.DataBaseType, id)

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

func Api_json(w http.ResponseWriter, r *http.Request) {

	//1
	CRM_Counter_Prometheus_JSON.Inc()

	if r.Method == "GET" {

		// // get parametrs from get-http
		// for key, value := range r.Header {
		// 	if key == "Token" {
		// 		fmt.Println("Token:" + value[0])
		// 	}
		// }

		customer_map_s := EngineCRMv.GetAllCustomer(EngineCRMv.DataBaseType)

		JsonString, err := json.Marshal(customer_map_s)
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, "error json:"+err.Error())
		}
		fmt.Fprintf(w, string(JsonString))

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

		for _, p := range customer_map_json {
			EngineCRMv.AddChangeOneRow(EngineCRMv.DataBaseType, p)
		}

		fmt.Fprintf(w, string(body))

	}

}

func Api_xml(w http.ResponseWriter, r *http.Request) {

	//1
	CRM_Counter_Prometheus_XML.Inc()

	if r.Method == "GET" {

		customer_map_s := EngineCRMv.GetAllCustomer(EngineCRMv.DataBaseType)

		doc := etree.NewDocument()
		//doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

		Custromers := doc.CreateElement("Custromers")

		for _, p := range customer_map_s {
			Custromer := Custromers.CreateElement("Custromer")
			Custromer.CreateAttr("value", p.Customer_id)

			id := Custromer.CreateElement("Customer_id")
			id.CreateAttr("value", p.Customer_id)
			name := Custromer.CreateElement("Customer_name")
			name.CreateAttr("value", p.Customer_name)
			type1 := Custromer.CreateElement("Customer_type")
			type1.CreateAttr("value", p.Customer_type)
			email := Custromer.CreateElement("Customer_email")
			email.CreateAttr("value", p.Customer_email)
		}

		//doc.CreateText("/xml")

		doc.Indent(2)
		XMLString, _ := doc.WriteToString()

		fmt.Fprintf(w, XMLString)

	} else {

		// test_rez_slice := []CustomerStruct_xml{}
		// //var test_rez []Customer_struct
		// if err := xml.Unmarshal(xmlData, &test_rez_slice); err != nil {
		// 	panic(err)
		// }
		// fmt.Println(test_rez_slice)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		body = []byte(`<Custromers>
		 <Custromer value="777">
		   <Customer_id value="777"/>
		   <Customer_name value="Dmitry"/>
		   <Customer_type value="Cust"/>
		   <Customer_email value="fff@mail.ru"/>
		 </Custromer>
		 <Custromer value="666">
		   <Customer_id value="666"/>
		   <Customer_name value="Alex"/>
		   <Customer_type value="Cust_Fiz"/>
		   <Customer_email value="44fish@mail.ru"/>
		 </Custromer>
		</Custromers>`)

		doc := etree.NewDocument()
		if err := doc.ReadFromBytes(body); err != nil {
			panic(err)
		}

		var customer_map_xml = make(map[string]Customer_struct)

		Custromers := doc.SelectElement("Custromers")

		for _, Custromer := range Custromers.SelectElements("Custromer") {

			Customer_struct := Customer_struct{}
			//fmt.Println("CHILD element:", Custromer.Tag)
			if Customer_id := Custromer.SelectElement("Customer_id"); Customer_id != nil {
				value := Customer_id.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_id = value
			}
			if Customer_name := Custromer.SelectElement("Customer_name"); Customer_name != nil {
				value := Customer_name.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_name = value
			}
			if Customer_type := Custromer.SelectElement("Customer_type"); Customer_type != nil {
				value := Customer_type.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_type = value
			}

			if Customer_email := Custromer.SelectElement("Customer_email"); Customer_email != nil {
				value := Customer_email.SelectAttrValue("value", "unknown")
				Customer_struct.Customer_email = value
			}

			customer_map_xml[Customer_struct.Customer_id] = Customer_struct
			// for _, attr := range Custromer.Attr {
			// 	fmt.Printf("  ATTR: %s=%s\n", attr.Key, attr.Value)
			// }
		}

		for _, p := range customer_map_xml {
			EngineCRMv.AddChangeOneRow(EngineCRMv.DataBaseType, p)
		}

		fmt.Fprintf(w, string(body))

	}
}

func initLog() {
	file, err := os.OpenFile("./logs/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
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

func initPrometheus() {
	CRM_Counter_Prometheus_JSON = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "CRM_Counter_JSON",
		})
	prometheus.MustRegister(CRM_Counter_Prometheus_JSON)

	CRM_Counter_Prometheus_XML = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "CRM_Counter_XML",
		})
	prometheus.MustRegister(CRM_Counter_Prometheus_XML)

	CRM_Counter_Gauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "CRM_Gauge",
		})
	prometheus.MustRegister(CRM_Counter_Gauge)
}

func initgRPC() {
	listener, err := net.Listen("tcp", ":5300")

	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterCRMswapServer(grpcServer, &server{})
	grpcServer.Serve(listener)
}

func main() {

	//fmt.Println(DBLocal.Test(5))

	Global_settingsV.LoadSettingsFromDisk()
	EngineCRMv.SetSettings(Global_settingsV)

	initLog()

	// type_memory_storage_flag := flag.String("type_memory_storage", "", "type storage data")
	// flag.Parse()

	// if *type_memory_storage_flag == "" {
	// 	type_memory_storage = "DemoRegime"

	// 	//temporary
	// 	type_memory_storage = "Redis"
	// 	EngineCRMv.SetDataBaseType("Redis")
	// } else {
	// 	type_memory_storage = *type_memory_storage_flag
	// 	EngineCRMv.SetDataBaseType(*type_memory_storage_flag)
	// }

	EngineCRMv.InitDataBase()
	defer EngineCRMv.databaseSQLite.Close()

	go initgRPC()

	router := mux.NewRouter()

	router.HandleFunc("/", indexHandler)

	//localhost:8181/user?name=Sam&age=21
	router.HandleFunc("/user", user)

	//localhost:8181/get_customer?customer_id="123"
	router.HandleFunc("/get_customer", get_customer)

	//http://localhost:8181/checkINN?customer_INN=7702807750
	router.HandleFunc("/checkINN", checkINN)

	router.HandleFunc("/add_change_customer", add_change_customer)
	router.HandleFunc("/postform_add_change_customer", postform_add_change_customer)

	router.HandleFunc("/list_customer", list_customer)

	router.HandleFunc("/mainpage", mainpage)

	// replace to HTTPS router
	router.HandleFunc("/login", RedirectToHTTPS)
	router.HandleFunc("/loginPost", RedirectToHTTPS)

	router.HandleFunc("/settings", settings)

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
	httpsMux.Handle("/", router_HTTPS)
	go http.ListenAndServeTLS(":8182", "./Cert/cert.pem", "./Cert/key.pem", httpsMux)

	initPrometheus()

	httpPrometheus := http.NewServeMux()
	httpPrometheus.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8183", httpPrometheus)

	http.Handle("/", router)
	http.ListenAndServe(":8181", nil)
	fmt.Println("Server is listening777...")

}
