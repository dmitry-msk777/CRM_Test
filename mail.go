package main

import (
	"fmt"

	"log"

	"time"

	enginecrm "github.com/dmitry-msk777/CRM_Test/enginecrm"
	handlers "github.com/dmitry-msk777/CRM_Test/handlers"

	graphglmy "github.com/dmitry-msk777/CRM_Test/graphgl"
	prometheus "github.com/dmitry-msk777/CRM_Test/prometheus"
	protomodul "github.com/dmitry-msk777/CRM_Test/protomodul"
	rootsctuct "github.com/dmitry-msk777/CRM_Test/rootdescription"
	_ "github.com/mattn/go-sqlite3"
)

func RabbitMQ_Consumer() {

	if enginecrm.EngineCRMv.RabbitMQ_channel == nil {
		//err := errors.New("Connection to RabbitMQ not established")
		//return err
		return
	}

	q, err := enginecrm.EngineCRMv.RabbitMQ_channel.QueueDeclare(
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

	msgs, err := enginecrm.EngineCRMv.RabbitMQ_channel.Consume(
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
			enginecrm.EngineCRMv.TestChan <- string(d.Body)
		}

	}
}

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
		go RabbitMQ_Consumer()
	}

	if enginecrm.EngineCRMv.Global_settings.UsePrometheus {
		prometheus.StartPrometheus()
	}

	graphglmy.StartGraphQL()

	handlers.InitHTTPSlogin()

	handlers.StratHandlers()
}
