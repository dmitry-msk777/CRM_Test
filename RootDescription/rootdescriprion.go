package rootdescription

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

var Global_settingsV Global_settings

var Cookie_CRMv Cookie_CRM

var Customer_map = make(map[string]Customer_struct)

var Cookiemap = make(map[string]string)

//var Users = make(map[string]string)

var Mass_settings = make([]string, 2)

const CookieName = "CookieCRM"

var LoggerCRMv LoggerCRM

type Customer_struct struct {
	Customer_id    string
	Customer_name  string
	Customer_type  string
	Customer_email string
	Address_Struct Address_Struct
}

type Address_Struct struct {
	Street string
	House  int
}

type Global_settings struct {
	DataBaseType       string
	Mail_smtpServer    string
	AddressMongoBD     string
	AddressRedis       string
	AddressRabbitMQ    string
	Mail_email         string
	Mail_password      string
	UseRabbitMQ        bool
	UsePrometheus      bool
	Dada_apiKey        string
	Dada_secretKey     string
	GORM_DataType      string
	GORM_ConnectString string
}

func (GlobalSettings *Global_settings) SaveSettingsOnDisk() {

	f, err := os.Create("./settings/config.json")
	if err != nil {
		log.Fatal(err)
	}

	JsonString, err := json.Marshal(GlobalSettings)
	if err != nil {
		//EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
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

	//Global_settingsV = Settings
	*GlobalSettings = Settings

	if err := file.Close(); err != nil {
		fmt.Println(err)
	}
}

type LoggerCRM struct {
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
}

func (LoggerCRM *LoggerCRM) InitLog() {

	file, err := os.OpenFile("./logs/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	LoggerCRM.InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	LoggerCRM.ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	LoggerCRM.ErrorLogger.Println("Starting the application...")
}

type Users_CRM struct {
	User     string
	Password string
}

type Cookie_CRM struct {
	Id   string
	User string
}

// convert in cookie_base type
func (Cookie_CRM *Cookie_CRM) GenerateId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

type ViewData struct {
	Title        string
	Message      string
	User         string
	DataBaseType string
	Customers    map[string]Customer_struct
}

type Log1C struct {
	Level string
	Date string
	ApplicationName string
	ApplicationPresentation string
	Event string
	EventPresentation string
	User string
	UserName string
	Computer string
	Metadata string
	MetadataPresentation string
	Comment string
	Data string
	DataPresentation string
	TransactionStatus string
	TransactionID string
	Connection string
	Session string
	ServerName string
	Port string
	SyncPort string
}
