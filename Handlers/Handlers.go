package Handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"net/mail"
	"net/smtp"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/olivere/elastic"
	"go.mongodb.org/mongo-driver/bson"

	EngineCRM "github.com/dmitry-msk777/CRM_Test/EngineCRM"
	Prometheus "github.com/dmitry-msk777/CRM_Test/Prometheus"
	RootSctuct "github.com/dmitry-msk777/CRM_Test/RootDescription"
	Utilities "github.com/dmitry-msk777/CRM_Test/Utilities"

	"encoding/json"
	"io/ioutil"

	"github.com/beevik/etree"
	"gopkg.in/webdeskltd/dadata.v2"
)

func List_customer(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM,
	PrometheusEngineV Prometheus.PrometheusEngine) {

	//prometheus
	if EngineCRMv.Global_settings.UsePrometheus {
		PrometheusEngineV.CRM_Counter_Gauge.Set(float64(5)) // or: Inc(), Dec(), Add(5), Dec(5),
	}

	tmpl, err := template.ParseFiles("templates/list_customer.html", "templates/header.html")
	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	customer_map_data, err := EngineCRMv.GetAllCustomer(EngineCRMv.DataBaseType)

	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl.ExecuteTemplate(w, "list_customer", customer_map_data)

}

func Services(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM) {

	tmpl, err := template.ParseFiles("templates/services.html", "templates/header.html")
	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl.ExecuteTemplate(w, "services", nil)

}

func Settings(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM) {

	if r.Method == "GET" {

		tmpl, err := template.ParseFiles("templates/settings.html", "templates/header.html")
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		tmpl.ExecuteTemplate(w, "settings", EngineCRMv.Global_settings)

	} else {

		//mass_settings[0] = r.FormValue("email")
		//mass_settings[1] = r.FormValue("password")

		EngineCRMv.Global_settings.Mail_email = r.FormValue("Mail_email")
		EngineCRMv.Global_settings.Mail_password = r.FormValue("Mail_password")
		EngineCRMv.Global_settings.Mail_smtpServer = r.FormValue("Mail_smtpServer")
		EngineCRMv.Global_settings.DataBaseType = r.FormValue("DataBaseType")

		EngineCRMv.Global_settings.AddressMongoBD = r.FormValue("AddressMongoBD")
		EngineCRMv.Global_settings.AddressRedis = r.FormValue("AddressRedis")
		EngineCRMv.Global_settings.AddressRabbitMQ = r.FormValue("AddressRabbitMQ")

		EngineCRMv.Global_settings.Dada_apiKey = r.FormValue("Dada_apiKey")
		EngineCRMv.Global_settings.Dada_secretKey = r.FormValue("Dada_secretKey")

		EngineCRMv.Global_settings.GORM_DataType = r.FormValue("GORM_DataType")
		EngineCRMv.Global_settings.GORM_ConnectString = r.FormValue("GORM_ConnectString")

		if r.FormValue("UseRabbitMQ") == "on" {
			EngineCRMv.Global_settings.UseRabbitMQ = true
		} else {
			EngineCRMv.Global_settings.UseRabbitMQ = false
		}

		if r.FormValue("UsePrometheus") == "on" {
			EngineCRMv.Global_settings.UsePrometheus = true
		} else {
			EngineCRMv.Global_settings.UsePrometheus = false
		}

		EngineCRMv.SetSettings(EngineCRMv.Global_settings)

		err := EngineCRMv.InitDataBase()
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		EngineCRMv.Global_settings.SaveSettingsOnDisk()

		http.Redirect(w, r, "/", 302)
	}
}

func EditHandler(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM, Global_settingsV RootSctuct.Global_settings) {
	err := r.ParseForm()
	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
	}

	Customer_struct_out := RootSctuct.Customer_struct{
		Customer_id:    r.FormValue("customer_id"),
		Customer_name:  r.FormValue("customer_name"),
		Customer_type:  r.FormValue("customer_type"),
		Customer_email: r.FormValue("customer_email"),
	}

	EngineCRMv.AddChangeOneRow(EngineCRMv.DataBaseType, Customer_struct_out, Global_settingsV)

	//return err
	//fmt.Fprintf(w, err.Error())

	http.Redirect(w, r, "/list_customer", 301)

}

func EditPage(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM,
	Global_settingsV RootSctuct.Global_settings, id string) {
	// vars := mux.Vars(r)
	// id := vars["id"]

	Customer_struct_out, err := EngineCRMv.FindOneRow(EngineCRMv.DataBaseType, id, Global_settingsV)

	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl, err := template.ParseFiles("templates/edit.html", "templates/header.html")
	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl.ExecuteTemplate(w, "edit", Customer_struct_out)

}

func DeleteHandler(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM,
	Global_settingsV RootSctuct.Global_settings, id string) {
	// vars := mux.Vars(r)
	// id := vars["id"]

	err := EngineCRMv.DeleteOneRow(EngineCRMv.DataBaseType, id, Global_settingsV)

	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	http.Redirect(w, r, "/list_customer", 301)

}

func Api_json(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM, PrometheusEngineV Prometheus.PrometheusEngine,
	Global_settingsV RootSctuct.Global_settings) {

	if EngineCRMv.Global_settings.UsePrometheus {
		//1
		PrometheusEngineV.CRM_Counter_Prometheus_JSON.Inc()
	}

	if r.Method == "GET" {

		// // get parametrs from get-http
		// for key, value := range r.Header {
		// 	if key == "Token" {
		// 		fmt.Println("Token:" + value[0])
		// 	}
		// }

		customer_map_s, err := EngineCRMv.GetAllCustomer(EngineCRMv.DataBaseType)

		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		JsonString, err := json.Marshal(customer_map_s)
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, "error json:"+err.Error())
		}
		fmt.Fprintf(w, string(JsonString))

	} else {

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		var customer_map_json = make(map[string]RootSctuct.Customer_struct)

		err = json.Unmarshal(body, &customer_map_json)
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		for _, p := range customer_map_json {
			err := EngineCRMv.AddChangeOneRow(EngineCRMv.DataBaseType, p, Global_settingsV)
			if err != nil {
				EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
				fmt.Println(err.Error())
			}
		}

		fmt.Fprintf(w, string(body))

	}

}

func Api_xml(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM, PrometheusEngineV Prometheus.PrometheusEngine,
	Global_settingsV RootSctuct.Global_settings) {

	if EngineCRMv.Global_settings.UsePrometheus {
		//1
		PrometheusEngineV.CRM_Counter_Prometheus_XML.Inc()
	}

	if r.Method == "GET" {

		customer_map_s, err := EngineCRMv.GetAllCustomer(EngineCRMv.DataBaseType)

		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

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
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
		}

		// body = []byte(`<Custromers>
		//  <Custromer value="777">
		//    <Customer_id value="777"/>
		//    <Customer_name value="Dmitry"/>
		//    <Customer_type value="Cust"/>
		//    <Customer_email value="fff@mail.ru"/>
		//  </Custromer>
		//  <Custromer value="666">
		//    <Customer_id value="666"/>
		//    <Customer_name value="Alex"/>
		//    <Customer_type value="Cust_Fiz"/>
		//    <Customer_email value="44fish@mail.ru"/>
		//  </Custromer>
		// </Custromers>`)

		doc := etree.NewDocument()
		if err := doc.ReadFromBytes(body); err != nil {
			panic(err)
		}

		var customer_map_xml = make(map[string]RootSctuct.Customer_struct)

		Custromers := doc.SelectElement("Custromers")

		for _, Custromer := range Custromers.SelectElements("Custromer") {

			Customer_struct := RootSctuct.Customer_struct{}
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
			err := EngineCRMv.AddChangeOneRow(EngineCRMv.DataBaseType, p, Global_settingsV)
			if err != nil {
				EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
				fmt.Println(err.Error())
			}
		}

		fmt.Fprintf(w, string(body))

	}
}

func SuggestAddresses(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM,
	Global_settingsV RootSctuct.Global_settings) {

	customer_Address := r.URL.Query().Get("customer_Address")

	// https://github.com/webdeskltd/dadata/blob/v2/examples_test.go

	// split the line with the address has become a paid function
	// daData := dadata.NewDaData("1aa37d40a1f8267955a88cb429e6bbdff3c33a31", "b8fb31a67d75d925755b04b754c514d3e2d9fe70")
	// //func ExampleDaData_CleanAddresses() {
	// addresses, err := daData.CleanAddresses("ул.Правды 26", "пер.Расковой 5")
	// if nil != err {
	// 	fmt.Println(err)
	// }
	// for _, address := range addresses {
	// 	fmt.Println(address.StreetTypeFull)
	// 	fmt.Println(address.Street)
	// 	fmt.Println(address.House)
	// }
	// // Output:
	// // улица
	// // Правды
	// // 26
	// // переулок
	// // Расковой
	// // 5
	// //}

	daData2 := dadata.NewDaData(EngineCRMv.Global_settings.Dada_apiKey, EngineCRMv.Global_settings.Dada_secretKey)

	addresses2, err := daData2.SuggestAddresses(dadata.SuggestRequestParams{Query: customer_Address, Count: 5})
	if nil != err {
		fmt.Fprintf(w, err.Error())
	}

	for _, address2 := range addresses2 {
		fmt.Fprintf(w, address2.UnrestrictedValue)
		fmt.Fprintf(w, address2.Data.Street)
		fmt.Fprintf(w, address2.Data.FiasLevel)
		fmt.Fprintln(w, "")
	}

	// Output:
	// г Москва, Пресненская наб
	// Пресненская
	// 7
	// г Москва, ул Пресненский Вал
	// Пресненский Вал
	// 7

	//}

}

func IndexHandler(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM,
	Global_settingsV RootSctuct.Global_settings, cookieName string, customer_map map[string]RootSctuct.Customer_struct) {

	t, err := template.ParseFiles("templates/main_page.html", "templates/header.html")
	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	nameUserFromCookieStruc := ""

	CookieGet, _ := r.Cookie(cookieName)
	if CookieGet != nil {
		nameUserFromCookie, flagmap := EngineCRMv.Cookie_CRM_map[CookieGet.Value]
		if flagmap != false {
			nameUserFromCookieStruc = nameUserFromCookie.User
		}
	}

	if EngineCRMv.DataBaseType == "SQLit" && CookieGet != nil {

		rows, err := EngineCRMv.DatabaseSQLite.Query("select * from cookie where id = $1", CookieGet.Value)
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			panic(err)
		}
		defer rows.Close()
		cookie_base_s := []RootSctuct.Cookie_CRM{}

		for rows.Next() {
			p := RootSctuct.Cookie_CRM{}
			err := rows.Scan(&p.Id, &p.User)
			if err != nil {
				EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
				fmt.Println(err)
				continue
			}
			cookie_base_s = append(cookie_base_s, p)
		}
		for _, p := range cookie_base_s {
			nameUserFromCookieStruc = p.User
			fmt.Println(p.Id, p.User)
		}

	}

	data := RootSctuct.ViewData{
		Title:     "list customer",
		Message:   "list customer below",
		User:      nameUserFromCookieStruc,
		Customers: customer_map,
	}

	// t.ExecuteTemplate(w, "main_page", customer_map)
	t.ExecuteTemplate(w, "main_page", data)
}

func CheckINN(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM,
	Global_settingsV RootSctuct.Global_settings) {

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
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
	}

	req.ContentLength = int64(len(soapQuery))

	req.Header.Add("Content-Type", "text/xml;charset=UTF-8")
	req.Header.Add("Accept", "text/xml")
	req.Header.Add("SOAPAction", "NdsRequest2")

	resp, err := client.Do(req)
	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
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

func Get_customer(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM,
	Global_settingsV RootSctuct.Global_settings, customer_map map[string]RootSctuct.Customer_struct) {

	customer_id_for_find := r.URL.Query().Get("customer_id")

	switch EngineCRMv.DataBaseType {
	case "SQLit":
		fmt.Fprintf(w, "function not implemented for SQLit")
	case "MongoDB":

		cur, err := EngineCRMv.CollectionMongoDB.Find(context.Background(), bson.D{})
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		}
		defer cur.Close(context.Background())

		Customer_struct_slice := []RootSctuct.Customer_struct{}

		for cur.Next(context.Background()) {

			Customer_struct_out := RootSctuct.Customer_struct{}

			err := cur.Decode(&Customer_struct_out)
			if err != nil {
				EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			}

			Customer_struct_slice = append(Customer_struct_slice, Customer_struct_out)

		}

		if err := cur.Err(); err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		}

		//ElasticSerch

		clientElasticSerch, err := elastic.NewClient(elastic.SetSniff(false),
			elastic.SetURL("http://127.0.0.1:32771", "http://127.0.0.1:32770"))
		// elastic.SetBasicAuth("user", "secret"))
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		exists, err := clientElasticSerch.IndexExists("crm_customer").Do(context.Background()) //twitter
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
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
				EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
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
				EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
				fmt.Fprintf(w, err.Error())
				return
			}
			fmt.Printf("Indexed tweet %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)

		}

		// Flush to make sure the documents got written.
		_, err = clientElasticSerch.Flush().Index("crm_customer").Do(context.Background())
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
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
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			fmt.Fprintf(w, err.Error())
			return
		}

		// searchResult is of type SearchResult and returns hits, suggestions,
		// and all kinds of other information from Elasticsearch.
		fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

		var ttyp RootSctuct.Customer_struct
		for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
			t := item.(RootSctuct.Customer_struct)
			fmt.Fprintf(w, "customer_id: %s customer_name: %s", t.Customer_id, t.Customer_name)
		}

		fmt.Printf("Found a total of %d tweets\n", searchResult.TotalHits())

		// Delete an index.
		deleteIndex, err := clientElasticSerch.DeleteIndex("crm_customer").Do(context.Background())
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
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

func Send_message(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM,
	Global_settingsV RootSctuct.Global_settings) {

	// Set up authentication information. https://yandex.ru/support/mail/mail-clients.html

	//smtpServer := "smtp.yandex.ru"
	smtpServer := EngineCRMv.Global_settings.Mail_smtpServer
	auth := smtp.PlainAuth(
		"",
		EngineCRMv.Global_settings.Mail_email,
		EngineCRMv.Global_settings.Mail_password,
		smtpServer,
	)

	from := mail.Address{"Test", EngineCRMv.Global_settings.Mail_email}
	to := mail.Address{"test2", "dima-irk35@mail.ru"}
	title := "Title"

	body := "body"

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = Utilities.EncodeRFC2047(title)
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
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprint(w, "error"+err.Error())
	} else {
		http.Redirect(w, r, "/", 302)
	}

}

func LoginPost(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM,
	Global_settingsV RootSctuct.Global_settings, Cookie_CRMv RootSctuct.Cookie_CRM, CookieName string) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if EngineCRMv.DataBaseType == "SQLit" {

		rows, err := EngineCRMv.DatabaseSQLite.Query("select * from users where user = $1 and password = $2", username, password)
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			panic(err)
		}
		defer rows.Close()
		users_base_s := []RootSctuct.Users_CRM{}

		for rows.Next() {
			p := RootSctuct.Users_CRM{}
			err := rows.Scan(&p.User, &p.Password)
			if err != nil {
				EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
				fmt.Println(err)
				continue
			}
			users_base_s = append(users_base_s, p)
		}
		for _, p := range users_base_s {
			fmt.Println(p.User, p.Password)
		}

	} else {

		user_password_struct, flagusers := EngineCRMv.Users_CRM_map[username]
		if flagusers == true {
			if user_password_struct.Password != password {
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

		result, err := EngineCRMv.DatabaseSQLite.Exec("insert into cookie (id, user) values ($1, $2)",
			idcookie, username)
		if err != nil {
			EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
			panic(err)
		}
		fmt.Println(result.LastInsertId()) // id последнего добавленного объекта
		fmt.Println(result.RowsAffected()) // количество добавленных строк

	} else {
		cookie_CRM_data := RootSctuct.Cookie_CRM{
			Id:   idcookie,
			User: username,
		}
		EngineCRMv.Cookie_CRM_map[idcookie] = cookie_CRM_data
	}

	cookieHttp := &http.Cookie{
		Name:    CookieName,
		Value:   idcookie,
		Expires: time.Now().Add(6 * time.Minute),
	}

	http.SetCookie(w, cookieHttp)

	//fmt.Fprint(w, username+" "+password)
	//http.Redirect(w, r, "/", 302)
	http.Redirect(w, r, "http://localhost:8181/",
		http.StatusMovedPermanently)
}

func Add_change_customer(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM,
	Global_settingsV RootSctuct.Global_settings) {

	tmpl, err := template.ParseFiles("templates/add_change_customer.html", "templates/header.html")
	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	tmpl.ExecuteTemplate(w, "add_change_customer", nil)

}

func Postform_add_change_customer(w http.ResponseWriter, r *http.Request, EngineCRMv EngineCRM.EngineCRM,
	Global_settingsV RootSctuct.Global_settings) {

	customer_data := RootSctuct.Customer_struct{
		Customer_name:  r.FormValue("customer_name"),
		Customer_id:    r.FormValue("customer_id"),
		Customer_type:  r.FormValue("customer_type"),
		Customer_email: r.FormValue("customer_email"),
	}

	err := EngineCRMv.AddChangeOneRow(EngineCRMv.DataBaseType, customer_data, Global_settingsV)

	if err != nil {
		EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	http.Redirect(w, r, "/list_customer", 302)
}
