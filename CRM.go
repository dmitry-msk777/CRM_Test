package main

import (
	//	"bytes"
	"context"
	//	"encoding/base64"
	"fmt"
	//"html/template"

	//	"io/ioutil"
	"log"
	"net"
	"net/http"

	//	"net/mail"
	//	"net/smtp"

	//	"reflect"

	//	"regexp"
	//	"strings"
	"time"

	EngineCRM "github.com/dmitry-msk777/CRM_Test/EngineCRM"
	Handlers "github.com/dmitry-msk777/CRM_Test/Handlers"
	Prometheus "github.com/dmitry-msk777/CRM_Test/Prometheus"
	RootSctuct "github.com/dmitry-msk777/CRM_Test/RootDescription"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	//	"github.com/olivere/elastic"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	//	"go.mongodb.org/mongo-driver/bson"

	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/mongo/readpref"

	//pb "../CRM_Test/proto"
	pb "github.com/dmitry-msk777/CRM_Test/proto"
	"github.com/friendsofgo/graphiql"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var EngineCRMv EngineCRM.EngineCRM

var Global_settingsV RootSctuct.Global_settings

//GraphQL
const Schema = `
type Customer_struct {
    Customer_id: String!
    Customer_name: String!
	Customer_type: String!
	Customer_email: String!
}
type Query {
    FindOneRow(Customer_id: String!): Customer_struct
}
schema {
    query: Query
}
`

type FindOneRow_Resolver struct {
	v *RootSctuct.Customer_struct
}

func (r *FindOneRow_Resolver) Customer_id() string    { return r.v.Customer_id }
func (r *FindOneRow_Resolver) Customer_name() string  { return r.v.Customer_name }
func (r *FindOneRow_Resolver) Customer_type() string  { return r.v.Customer_type }
func (r *FindOneRow_Resolver) Customer_email() string { return r.v.Customer_email }

func (q *query) FindOneRow(ctx context.Context, args struct{ Customer_id string }) *FindOneRow_Resolver {

	v, err := EngineCRMv.FindOneRow(EngineCRMv.DataBaseType, args.Customer_id, Global_settingsV)

	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		return nil
	}

	return &FindOneRow_Resolver{v: &v}
	//return &v
}

type query struct{}

//GraphQL end

var Cookie_CRMv RootSctuct.Cookie_CRM

var PrometheusEngineV Prometheus.PrometheusEngine

var customer_map = make(map[string]RootSctuct.Customer_struct)

var cookiemap = make(map[string]string)
var users = make(map[string]string)

var mass_settings = make([]string, 2)

const CookieName = "CookieCRM"

var LoggerCRMv RootSctuct.LoggerCRM

//Protobuff
type server struct{}

func (s *server) GET_List(ctx context.Context, in *pb.RequestGET) (*pb.ResponseGET, error) {

	id := in.CustomerId

	Customer_struct_out, err := EngineCRMv.FindOneRow(EngineCRMv.DataBaseType, id, Global_settingsV)

	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		return nil, nil
	}

	response := &pb.ResponseGET{
		CustomerId:    Customer_struct_out.Customer_id,
		CustomerName:  Customer_struct_out.Customer_name,
		CustomerType:  Customer_struct_out.Customer_type,
		CustomerEmail: Customer_struct_out.Customer_email,
	}

	return response, nil
}

func (s *server) POST_List(ctx context.Context, in *pb.RequestPOST) (*pb.ResponsePOST, error) {

	Customer_struct_out := RootSctuct.Customer_struct{
		Customer_id:    in.CustomerId,
		Customer_name:  in.CustomerName,
		Customer_type:  in.CustomerType,
		Customer_email: in.CustomerEmail,
	}

	err := EngineCRMv.AddChangeOneRow(EngineCRMv.DataBaseType, Customer_struct_out, Global_settingsV)

	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		return nil, err
	}

	return &pb.ResponsePOST{CustomerId: "True"}, nil
}

//Protobuff end

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	Handlers.IndexHandler(w, r, EngineCRMv, Global_settingsV, CookieName, customer_map)
}

func Add_change_customer(w http.ResponseWriter, r *http.Request) {
	Handlers.Add_change_customer(w, r, EngineCRMv, Global_settingsV)
}

func Postform_add_change_customer(w http.ResponseWriter, r *http.Request) {
	Handlers.Postform_add_change_customer(w, r, EngineCRMv, Global_settingsV)
}

func List_customer(w http.ResponseWriter, r *http.Request) {
	Handlers.List_customer(w, r, EngineCRMv, PrometheusEngineV)
}

func Get_customer(w http.ResponseWriter, r *http.Request) {
	Handlers.Get_customer(w, r, EngineCRMv, Global_settingsV, customer_map)
}

func Services(w http.ResponseWriter, r *http.Request) {
	Handlers.Services(w, r, EngineCRMv)
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	Handlers.LoginPost(w, r, EngineCRMv, Global_settingsV, Cookie_CRMv, CookieName)
}

func RedirectToHTTPS(w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, "https://localhost:8182"+r.RequestURI,
		http.StatusMovedPermanently)

}

func login(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./login/login.html")
}

func Settings(w http.ResponseWriter, r *http.Request) {
	Handlers.Settings(w, r, EngineCRMv)
}

func Send_message(w http.ResponseWriter, r *http.Request) {
	Handlers.Send_message(w, r, EngineCRMv, Global_settingsV)
}

func EditPage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	Handlers.EditPage(w, r, EngineCRMv, Global_settingsV, id)

}

func EditHandler(w http.ResponseWriter, r *http.Request) {
	Handlers.EditHandler(w, r, EngineCRMv, Global_settingsV)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]

	Handlers.DeleteHandler(w, r, EngineCRMv, Global_settingsV, id)

}

func SuggestAddresses(w http.ResponseWriter, r *http.Request) {
	Handlers.SuggestAddresses(w, r, EngineCRMv, Global_settingsV)
}

func Api_json(w http.ResponseWriter, r *http.Request) {
	Handlers.Api_json(w, r, EngineCRMv, PrometheusEngineV, Global_settingsV)
}

func Api_xml(w http.ResponseWriter, r *http.Request) {
	Handlers.Api_xml(w, r, EngineCRMv, PrometheusEngineV, Global_settingsV)
}

func CheckINN(w http.ResponseWriter, r *http.Request) {
	Handlers.CheckINN(w, r, EngineCRMv, Global_settingsV)
}

func Test_handler(w http.ResponseWriter, r *http.Request) {
	//gorm.Test_gorm()
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

func RabbitMQ_Consumer() {

	if EngineCRMv.RabbitMQ_channel == nil {
		//err := errors.New("Connection to RabbitMQ not established")
		//return err
		return
	}

	q, err := EngineCRMv.RabbitMQ_channel.QueueDeclare(
		"Customer___add_change", // name
		false,                   // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)

	if err != nil {
		fmt.Println("Failed to declare a queue: ", err)
	}

	msgs, err := EngineCRMv.RabbitMQ_channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		fmt.Println("Failed to register a consumer: ", err)
	}

	for {
		time.Sleep(100 * time.Millisecond)
		//fmt.Println("123")

		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			EngineCRMv.TestChan <- string(d.Body)
		}

	}
}

func Test_Chan() {
	for {
		time.Sleep(100 * time.Millisecond)
		//fmt.Println("1234")
		//log.Printf("Chan consisit:", <-EngineCRMv.testChan)
		fmt.Println("Chan consisit:", <-EngineCRMv.TestChan)
	}
}

func main() {

	//fmt.Println(DBLocal.Test(5))

	Global_settingsV.LoadSettingsFromDisk()
	EngineCRMv.SetSettings(Global_settingsV)

	LoggerCRMv.InitLog()
	EngineCRMv.LoggerCRM = LoggerCRMv

	err := EngineCRMv.InitDataBase()
	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Println(err.Error())
		return
	}
	defer EngineCRMv.DatabaseSQLite.Close()

	go initgRPC()

	go Test_Chan()

	if EngineCRMv.Global_settings.UseRabbitMQ {
		EngineCRMv.InitRabbitMQ(Global_settingsV)
		go RabbitMQ_Consumer()
	}

	router := mux.NewRouter()

	router.HandleFunc("/", IndexHandler)

	//localhost:8181/get_customer?customer_id="123"
	router.HandleFunc("/get_customer", Get_customer)

	//http://localhost:8181/checkINN?customer_INN=7702807750
	router.HandleFunc("/checkINN", CheckINN)

	//http://localhost:8181/SuggestAddresses?customer_Address=Рязанский
	router.HandleFunc("/SuggestAddresses", SuggestAddresses)

	router.HandleFunc("/add_change_customer", Add_change_customer)
	router.HandleFunc("/postform_add_change_customer", Postform_add_change_customer)

	router.HandleFunc("/list_customer", List_customer)

	router.HandleFunc("/test", Test_handler)

	// replace to HTTPS router
	router.HandleFunc("/login", RedirectToHTTPS)
	router.HandleFunc("/loginPost", RedirectToHTTPS)

	router.HandleFunc("/settings", Settings)
	router.HandleFunc("/services", Services)

	router.HandleFunc("/send_message", Send_message)

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
	router_HTTPS.HandleFunc("/loginPost", LoginPost)

	httpsMux := http.NewServeMux()
	httpsMux.Handle("/", router_HTTPS)
	go http.ListenAndServeTLS(":8182", "./Cert/cert.pem", "./Cert/key.pem", httpsMux)

	if EngineCRMv.Global_settings.UsePrometheus {
		PrometheusEngineV.InitPrometheus()

		httpPrometheus := http.NewServeMux()
		httpPrometheus.Handle("/metrics", promhttp.Handler())
		go http.ListenAndServe(":8183", httpPrometheus)
	}

	//GraphQL
	httpGraphQL := http.NewServeMux()

	schema := graphql.MustParseSchema(Schema, &query{})
	httpGraphQL.Handle("/query", &relay.Handler{Schema: schema})

	// First argument must be same as graphql handler path
	graphiqlHandler, err := graphiql.NewGraphiqlHandler("/query")
	if err != nil {
		panic(err)
	}
	httpGraphQL.Handle("/", graphiqlHandler)

	go http.ListenAndServe(":8184", httpGraphQL)
	//GraphQL end

	http.Handle("/", router)
	http.ListenAndServe(":8181", nil)
	fmt.Println("Server is listening...")

}
