package rootdescription

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

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
