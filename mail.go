package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"time"

	enginecrm "github.com/dmitry-msk777/CRM_Test/enginecrm"
	handlers "github.com/dmitry-msk777/CRM_Test/handlers"

	consul "github.com/dmitry-msk777/CRM_Test/consul"
	graphglmy "github.com/dmitry-msk777/CRM_Test/graphgl"
	prometheus "github.com/dmitry-msk777/CRM_Test/prometheus"
	protomodul "github.com/dmitry-msk777/CRM_Test/protomodul"
	rootsctuct "github.com/dmitry-msk777/CRM_Test/rootdescription"
	_ "github.com/mattn/go-sqlite3"

	"github.com/nsqio/go-nsq"
)

func Test_Chan() {
	for {
		time.Sleep(100 * time.Millisecond)
		//fmt.Println("1234")
		//log.Printf("Chan consisit:", <-EngineCRMv.testChan)
		fmt.Println("Chan consisit:", <-enginecrm.EngineCRMv.TestChan)
	}
}

type myMessageHandler struct{}

// HandleMessage implements the Handler interface.
func (h *myMessageHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	// do whatever actual message processing is desired
	//err := processMessage(m.Body)

	fmt.Println("NSQ message:", string(m.Body))

	//var err error

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return nil
}

func TestConsumerNSQ() {

	// Instantiate a consumer that will subscribe to the provided channel.
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer("Topic_Test_1C", "Channel_Test_1C", config)
	if err != nil {
		log.Fatal(err)
	}

	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	consumer.AddHandler(&myMessageHandler{})

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	// localhost:4161
	err = consumer.ConnectToNSQLookupd("localhost:32853")
	if err != nil {
		log.Fatal(err)
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Gracefully stop the consumer.
	consumer.Stop()

}

func TestProduserNSQ() {

	config := nsq.NewConfig()
	// producer, err := nsq.NewProducer("127.0.0.1:4150", config)
	producer, err := nsq.NewProducer("Localhost:32857", config)
	if err != nil {
		log.Fatal(err)
	}

	inc := 0

	for {

		//time.Sleep(3000 * time.Millisecond)
		inc++

		if inc == 100000 {
			break
		}

		// Instantiate a producer.
		messageBody := []byte("hello5")
		topicName := "Topic_Test_1C"

		// Synchronously publish a single message to the specified topic.
		// Messages can also be sent asynchronously and/or in batches.
		err = producer.Publish(topicName, messageBody)
		if err != nil {
			log.Fatal(err)
		}

	}

	// Gracefully stop the producer when appropriate (e.g. before shutting down the service)
	producer.Stop()
}

func main() {

	// for function queryLookupd() in consumer.go module remove hardcoded about adress and pord,
	//go TestConsumerNSQ()
	//go TestProduserNSQ()

	consulAddr := flag.String("consulAddr", "localhost:32769", "a string")
	consulActive := flag.Bool("consulactive", false, "a bool")
	flag.Parse()
	//fmt.Println("Consul addres: ", *consulAddr)

	if *consulActive == false {
		rootsctuct.Global_settingsV = consul.GetSettingsFromConsul(*consulAddr)
	} else {
		rootsctuct.Global_settingsV.LoadSettingsFromDisk()
	}

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
