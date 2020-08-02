package enginecrm

import (
	"database/sql"
	"errors"
	"fmt"

	"time"

	"github.com/go-redis/redis/v7"

	"github.com/streadway/amqp"

	"encoding/json"

	"context"

	//DBLocal "./bd" //add extermal go module.
	DBLocal "github.com/dmitry-msk777/CRM_Test/bd"

	RootSctuct "github.com/dmitry-msk777/CRM_Test/RootDescription"

	_ "github.com/mattn/go-sqlite3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	gorm "github.com/dmitry-msk777/CRM_Test/gorm"
)

type EngineCRM struct {
	DataBaseType      string
	CollectionMongoDB *mongo.Collection
	DemoDBmap         map[string]RootSctuct.Customer_struct
	DatabaseSQLite    *sql.DB
	RedisClient       *redis.Client
	RabbitMQ_channel  *amqp.Channel
	Global_settings   RootSctuct.Global_settings
	Users_CRM_map     map[string]RootSctuct.Users_CRM
	Cookie_CRM_map    map[string]RootSctuct.Cookie_CRM
	TestChan          chan string
	LoggerCRM         RootSctuct.LoggerCRM
}

func (EngineCRM *EngineCRM) SetSettings(Global_settings RootSctuct.Global_settings) {

	EngineCRM.DataBaseType = Global_settings.DataBaseType
	if Global_settings.DataBaseType == "" {
		EngineCRM.DataBaseType = "DemoRegime"
		Global_settings.DataBaseType = "DemoRegime"
	}

	//EngineCRM.Global_settings = Global_settingsV
	EngineCRM.Global_settings = Global_settings

}

// Test. delete
func (EngineCRM *EngineCRM) GetOneJSON(a interface{}) (interface{}, interface{}) {
	JsonString, err := json.Marshal(EngineCRM.DemoDBmap)
	if err != nil {
		return err.Error(), JsonString
	}

	return string(JsonString), JsonString
}

func (EngineCRM *EngineCRM) InitDataBase() error {

	switch EngineCRM.DataBaseType {
	case "SQLit":
		db, err := sql.Open("sqlite3", "./bd/SQLit/base_sqlit.db")

		if err != nil {
			return err
		}
		//database = db
		EngineCRM.DatabaseSQLite = db

		// CREATE TABLE "customer" (
		// 	"customer_id"	TEXT NOT NULL,
		// 	"customer_name"	TEXT,
		// 	"customer_type"	TEXT,
		// 	"customer_email"	TEXT,
		// 	PRIMARY KEY("customer_id")
		// );
		sql_query := "create table if not exists customer (customer_id text primary key, customer_name text, customer_type text, customer_email text);"
		_, err = EngineCRM.DatabaseSQLite.Exec(sql_query)
		if err != nil {
			fmt.Println("can't create table : " + err.Error())
			return err
		}

		// CREATE TABLE "cookie" (
		// 	"id"	TEXT NOT NULL,
		// 	"user"	TEXT,
		// 	PRIMARY KEY("id")
		// );
		sql_query = "create table if not exists cookie (id text primary key, user text);"
		_, err = EngineCRM.DatabaseSQLite.Exec(sql_query)
		if err != nil {
			fmt.Println("can't create table : " + err.Error())
			return err
		}

		// CREATE TABLE "users" (
		// 	"user"	TEXT NOT NULL,
		// 	"password"	TEXT,
		// 	PRIMARY KEY("user")
		// );
		sql_query = "create table if not exists users (user text primary key, password text);"
		_, err = EngineCRM.DatabaseSQLite.Exec(sql_query)
		if err != nil {
			fmt.Println("can't create table : " + err.Error())
			return err
		}

	case "Redis":
		//localhost:32769
		EngineCRM.RedisClient = intiRedisClient(EngineCRM.Global_settings.AddressRedis)

		pong, err := EngineCRM.RedisClient.Ping().Result()
		if err != nil {
			EngineCRM.RedisClient = nil
			fmt.Println(pong, err)
			return err
		}

	case "MongoDB":

		//temporary
		//collectionMongoDB = GetCollectionMongoBD("CRM", "customers", "mongodb://localhost:32768")
		//"mongodb://localhost:32768"
		EngineCRM.CollectionMongoDB = DBLocal.GetCollectionMongoBD("CRM", "customers", EngineCRM.Global_settings.AddressMongoBD)

	default:

		var ArrayCustomer []RootSctuct.Customer_struct

		ArrayCustomer = append(ArrayCustomer, RootSctuct.Customer_struct{
			Customer_id:    "777",
			Customer_name:  "Dmitry",
			Customer_type:  "Cust",
			Customer_email: "fff@mail.ru",
		})

		ArrayCustomer = append(ArrayCustomer, RootSctuct.Customer_struct{
			Customer_id:    "666",
			Customer_name:  "Alex",
			Customer_type:  "Cust_Fiz",
			Customer_email: "44fish@mail.ru",
		})

		var mapForEngineCRM = make(map[string]RootSctuct.Customer_struct)
		EngineCRM.DemoDBmap = mapForEngineCRM

		var users_CRM_def = make(map[string]RootSctuct.Users_CRM)
		EngineCRM.Users_CRM_map = users_CRM_def

		users_CRM_data := RootSctuct.Users_CRM{
			User:     "admin",
			Password: "1234"}

		EngineCRM.Users_CRM_map["admin"] = users_CRM_data

		var cookie_CRM_def = make(map[string]RootSctuct.Cookie_CRM)
		EngineCRM.Cookie_CRM_map = cookie_CRM_def

		for _, p := range ArrayCustomer {
			EngineCRM.DemoDBmap[p.Customer_id] = p
		}

		EngineCRM.TestChan = make(chan string)

	}

	return nil
}

func (EngineCRM *EngineCRM) GetAllCustomer(DataBaseType string) (map[string]RootSctuct.Customer_struct, error) {

	var customer_map_s = make(map[string]RootSctuct.Customer_struct)

	switch DataBaseType {
	case "SQLit":

		rows, err := EngineCRM.DatabaseSQLite.Query("select * from customer")
		if err != nil {
			return customer_map_s, err
		}
		defer rows.Close()
		Customer_struct_s := []RootSctuct.Customer_struct{}

		for rows.Next() {
			p := RootSctuct.Customer_struct{}
			err := rows.Scan(&p.Customer_id, &p.Customer_name, &p.Customer_type, &p.Customer_email)
			if err != nil {
				EngineCRM.LoggerCRM.ErrorLogger.Println(err.Error())
				fmt.Println(err)
				continue
			}
			Customer_struct_s = append(Customer_struct_s, p)
		}
		for _, p := range Customer_struct_s {
			customer_map_s[p.Customer_id] = p
		}

		return customer_map_s, nil

	case "MongoDB":

		cur, err := EngineCRM.CollectionMongoDB.Find(context.Background(), bson.D{})
		if err != nil {
			return customer_map_s, err
		}
		defer cur.Close(context.Background())

		Customer_struct_slice := []RootSctuct.Customer_struct{}

		for cur.Next(context.Background()) {

			Customer_struct_out := RootSctuct.Customer_struct{}

			err := cur.Decode(&Customer_struct_out)
			if err != nil {
				return customer_map_s, err
			}

			Customer_struct_slice = append(Customer_struct_slice, Customer_struct_out)

			// To get the raw bson bytes use cursor.Current
			// // raw := cur.Current
			// // fmt.Println(raw)
			// do something with raw...
		}
		if err := cur.Err(); err != nil {
			return customer_map_s, err
		}

		for _, p := range Customer_struct_slice {
			customer_map_s[p.Customer_id] = p
		}

		return customer_map_s, nil

	case "Redis":

		var cursor uint64
		ScanCmd := EngineCRM.RedisClient.Scan(cursor, "", 100)
		//fmt.Println(ScanCmd)

		cursor1, _, err := ScanCmd.Result()

		if err != nil {
			EngineCRM.LoggerCRM.ErrorLogger.Println("key2 does not exist")
			return customer_map_s, err
		}

		//fmt.Println(cursor1, keys1)

		Customer_struct_slice := []RootSctuct.Customer_struct{}
		for _, value := range cursor1 {
			p := RootSctuct.Customer_struct{}
			//IDString := strconv.FormatInt(int64(i), 10)
			val2, err := EngineCRM.RedisClient.Get(value).Result()
			if err == redis.Nil {
				EngineCRM.LoggerCRM.ErrorLogger.Println("key2 does not exist")
				continue
				//fmt.Println("key2 does not exist")
			} else if err != nil {
				EngineCRM.LoggerCRM.ErrorLogger.Println(err.Error())
				continue
			} else {
				//fmt.Println("key2", val2)

				err = json.Unmarshal([]byte(val2), &p)
				if err != nil {
					EngineCRM.LoggerCRM.ErrorLogger.Println(err.Error())
					continue
				}

				Customer_struct_slice = append(Customer_struct_slice, p)
			}
		}

		for _, p := range Customer_struct_slice {
			customer_map_s[p.Customer_id] = p
		}

		return customer_map_s, nil

	case "gORM":

		Customer_struct_slice_gorm, err := gorm.GetAllCustomer(EngineCRM.Global_settings)
		if err != nil {
			return customer_map_s, err
		}

		for _, p := range Customer_struct_slice_gorm {
			customer_map_s[p.Customer_id] = p
		}

		return customer_map_s, nil

	default:
		return EngineCRM.DemoDBmap, nil
	}

}

func (EngineCRM *EngineCRM) FindOneRow(DataBaseType string, id string, Global_settings RootSctuct.Global_settings) (RootSctuct.Customer_struct, error) {

	Customer_struct_out := RootSctuct.Customer_struct{}

	switch DataBaseType {
	case "SQLit":

		row := EngineCRM.DatabaseSQLite.QueryRow("select * from customer where customer_id = ?", id)

		err := row.Scan(&Customer_struct_out.Customer_id, &Customer_struct_out.Customer_name, &Customer_struct_out.Customer_type, &Customer_struct_out.Customer_email)
		if err != nil {
			return Customer_struct_out, err
		}

	case "MongoDB":

		err := EngineCRM.CollectionMongoDB.FindOne(context.TODO(), bson.D{{"customer_id", id}}).Decode(&Customer_struct_out)
		if err != nil {
			// ErrNoDocuments means that the filter did not match any documents in the collection
			if err == mongo.ErrNoDocuments {
				return Customer_struct_out, err
			}
		}
		fmt.Printf("found document %v", Customer_struct_out)

	case "Redis":

		val2, err := EngineCRM.RedisClient.Get(id).Result()
		if err == redis.Nil {
			EngineCRM.LoggerCRM.ErrorLogger.Println("key2 does not exist")
			return Customer_struct_out, err
		} else if err != nil {
			EngineCRM.LoggerCRM.ErrorLogger.Println(err.Error())
			return Customer_struct_out, err
		} else {
			err = json.Unmarshal([]byte(val2), &Customer_struct_out)
			if err != nil {
				EngineCRM.LoggerCRM.ErrorLogger.Println(err.Error())
				return Customer_struct_out, err
			}

			return Customer_struct_out, nil
		}

	case "gORM":

		Customer_struct, err := gorm.FindOneRow(id, Global_settings)

		Customer_struct_out = Customer_struct

		if err != nil {
			return Customer_struct_out, nil
		}

	default:
		Customer_struct_out = EngineCRM.DemoDBmap[id]
	}

	return Customer_struct_out, nil
}

func (EngineCRM *EngineCRM) AddChangeOneRow(DataBaseType string, Customer_struct RootSctuct.Customer_struct, Global_settings RootSctuct.Global_settings) error {

	switch DataBaseType {
	case "SQLit":

		var count int

		row := EngineCRM.DatabaseSQLite.QueryRow("select COUNT(*) from customer where customer_id = ?", Customer_struct.Customer_id)

		err := row.Scan(&count)
		if err != nil {
			return err
		}

		if count == 0 {

			_, err = EngineCRM.DatabaseSQLite.Exec("insert into customer (customer_id, customer_name, customer_type, customer_email) values (?, ?, ?, ?)",
				Customer_struct.Customer_id, Customer_struct.Customer_name, Customer_struct.Customer_type, Customer_struct.Customer_email)

			if err != nil {
				return err
			}
		} else {
			_, err = EngineCRM.DatabaseSQLite.Exec("update customer set customer_name=?, customer_type=?, customer_email=? where customer_id=?",
				Customer_struct.Customer_name, Customer_struct.Customer_type, Customer_struct.Customer_email, Customer_struct.Customer_id)

			if err != nil {
				return err
			}
		}

	case "MongoDB":

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

		//maybe use? insertMany(): добавляет несколько документов
		//before adding find db.users.find()

		insertResult, err := EngineCRM.CollectionMongoDB.InsertOne(ctx, Customer_struct)
		if err != nil {
			return err
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
			return err
		}

		err = EngineCRM.RedisClient.Set(Customer_struct.Customer_id, string(JsonStr), 0).Err()
		if err != nil {
			return err
		}

	case "gORM":

		err := gorm.AddChangeOneRow(Customer_struct, Global_settings)

		if err != nil {
			return err
		}

	default:
		EngineCRM.DemoDBmap[Customer_struct.Customer_id] = Customer_struct
	}

	if EngineCRM.Global_settings.UseRabbitMQ {
		EngineCRM.SendInQueue(Customer_struct)
	}

	return nil
}

func (EngineCRM *EngineCRM) DeleteOneRow(DataBaseType string, id string, Global_settings RootSctuct.Global_settings) error {

	switch DataBaseType {
	case "SQLit":
		_, err := EngineCRM.DatabaseSQLite.Exec("delete from customer where customer_id = ?", id)
		if err != nil {
			return err
		}
	case "MongoDB":

		res, err := EngineCRM.CollectionMongoDB.DeleteOne(context.TODO(), bson.D{{"customer_id", id}})
		if err != nil {
			return err
		}
		fmt.Printf("deleted %v documents\n", res.DeletedCount)

	case "Redis":

		//iter := EngineCRMv.RedisClient.Scan(0, "prefix*", 0).Iterator()
		iter := EngineCRM.RedisClient.Scan(0, id, 0).Iterator()
		for iter.Next() {
			err := EngineCRM.RedisClient.Del(iter.Val()).Err()
			if err != nil {
				EngineCRM.LoggerCRM.ErrorLogger.Println(err.Error())
				return err
			}
			//fmt.Println(iter.Val())
		}
		if err := iter.Err(); err != nil {
			EngineCRM.LoggerCRM.ErrorLogger.Println(err.Error())
			return err
		}

	case "gORM":

		err := gorm.DeleteOneRow(id, Global_settings)

		if err != nil {
			return err
		}

	default:
		_, ok := EngineCRM.DemoDBmap[id]
		if ok {
			delete(EngineCRM.DemoDBmap, id)
		}
	}

	return nil

}

func (EngineCRM *EngineCRM) SendInQueue(Customer_struct RootSctuct.Customer_struct) error {

	if EngineCRM.RabbitMQ_channel == nil {
		err := errors.New("Connection to RabbitMQ not established")
		return err
	}

	q, err := EngineCRM.RabbitMQ_channel.QueueDeclare(
		"Customer___add_change", // name
		false,                   // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		return err
	}

	bodyJSON, err := json.Marshal(Customer_struct)
	if err != nil {
		return err
	}

	err = EngineCRM.RabbitMQ_channel.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        bodyJSON,
		})

	if err != nil {
		return err
	}

	return nil

}

func (EngineCRM *EngineCRM) InitRabbitMQ(Global_settings RootSctuct.Global_settings) error {

	// Experimenting with RabbitMQ on your workstation? Try the community Docker image:
	// docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

	conn, err := amqp.Dial(Global_settings.AddressRabbitMQ) //5672
	if err != nil {
		EngineCRM.LoggerCRM.ErrorLogger.Println("Failed to connect to RabbitMQ")
		return err
	}
	//defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		EngineCRM.LoggerCRM.ErrorLogger.Println("Failed to open a channel")
		return err
	}
	//defer ch.Close()

	EngineCRM.RabbitMQ_channel = ch

	return nil
}

func intiRedisClient(Addr string) *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return client
}
