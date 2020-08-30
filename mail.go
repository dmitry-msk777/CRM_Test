package main

import (
	"fmt"

	"time"

	enginecrm "github.com/dmitry-msk777/CRM_Test/enginecrm"
	handlers "github.com/dmitry-msk777/CRM_Test/handlers"

	graphglmy "github.com/dmitry-msk777/CRM_Test/graphgl"
	prometheus "github.com/dmitry-msk777/CRM_Test/prometheus"
	protomodul "github.com/dmitry-msk777/CRM_Test/protomodul"
	rootsctuct "github.com/dmitry-msk777/CRM_Test/rootdescription"
	_ "github.com/mattn/go-sqlite3"
)

func Test_Chan() {
	for {
		time.Sleep(100 * time.Millisecond)
		//fmt.Println("1234")
		//log.Printf("Chan consisit:", <-EngineCRMv.testChan)
		fmt.Println("Chan consisit:", <-enginecrm.EngineCRMv.TestChan)
	}
}

func main() {

	rootsctuct.Global_settingsV.LoadSettingsFromDisk()
	enginecrm.EngineCRMv.SetSettings(rootsctuct.Global_settingsV)

	rootsctuct.LoggerCRMv.InitLog()
	enginecrm.EngineCRMv.LoggerCRM = rootsctuct.LoggerCRMv

	err := enginecrm.EngineCRMv.InitDataBase()
	if err != nil {
		enginecrm.EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		fmt.Println(err.Error())
		return
	}
	//defer EngineCRMv.DatabaseSQLite.Close()

	go protomodul.InitgRPC()

	go Test_Chan()

	if enginecrm.EngineCRMv.Global_settings.UseRabbitMQ {
		enginecrm.EngineCRMv.InitRabbitMQ(rootsctuct.Global_settingsV)
		//go RabbitMQ_Consumer()
	}

	if enginecrm.EngineCRMv.Global_settings.UsePrometheus {
		prometheus.StartPrometheus()
	}

	graphglmy.StartGraphQL()

	handlers.InitHTTPSlogin()

	handlers.StratHandlers()
}
