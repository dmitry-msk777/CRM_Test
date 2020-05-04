//package testing
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	// how import root package CRM.go ???
	//DB "../bd"
	DBLocalTest "../"
	DBLocal "../bd"
)

var counter int

var collectionMongoDB *mongo.Collection

func Write_in_MongoBD(i int) {

	fmt.Println(&DBLocalTest.Customer_struct)

	Customer_struct_MDB := &DBLocal.Customer_struct{
		Customer_name:  "Customer_name",
		Customer_id:    "Customer_id",
		Customer_type:  "Customer_type",
		Customer_email: "Customer_email"}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	insertResult, err := collectionMongoDB.InsertOne(ctx, Customer_struct_MDB)
	if err != nil {
		panic(err)
	}
	fmt.Println(insertResult.InsertedID)

	counter++

}

func Just_test(i int) {
	counter++
	fmt.Println(counter, i)
}

func Call_http_get_test(i int) {
	resp, err := http.Get("http://localhost:8181/api_json")
	if err != nil {
		fmt.Println(err.Error)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(i)
	fmt.Println(string(body))
	counter++
}

func main() {
	fmt.Println("start testing")
	//fmt.Println(DBLocal.Test777(5))

	Customer_struct_MDB := &DBLocalTest.Customer_struct{
		Customer_name:  "Customer_name",
		Customer_id:    "Customer_id",
		Customer_type:  "Customer_type",
		Customer_email: "Customer_email"}

	fmt.Println(Customer_struct_MDB)

	counter = 666

	collectionMongoDB = DBLocal.GetCollectionMongoBD("CRM", "testing", "mongodb://localhost:32768")

	for i := 1; i <= 1000; i++ {
		//fmt.Println(":", i)
		//go Call_http_get_test(i)
		//Call_http_get_test(i)

		//go Just_test(i)
		go Write_in_MongoBD(i)
	}

	fmt.Println("end testing ", counter)

	//fmt.Println(DB.Test(5))

}
