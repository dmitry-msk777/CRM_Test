package main

import (
	"context"
	"log"
	"time"

	//pb "../proto"
	pb "github.com/dmitry-msk777/CRM_Test/proto"
	"google.golang.org/grpc"
)

func main() {

	conn, err := grpc.Dial("localhost:5300", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewCRMswapClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	//Retrive
	RequestGET := &pb.RequestGET{
		CustomerId: "2323111",
	}

	rply, err := c.GET_List(ctx, RequestGET)
	if err != nil {
		log.Println("something went wrong", err)
	}
	log.Println(rply)

	// // Record
	// RequestPOST := &pb.RequestPOST{
	// 	CustomerId:    "123007",
	// 	CustomerName:  "Alex gRPC",
	// 	CustomerType:  "Google",
	// 	CustomerEmail: "gRPC@mail.ru",
	// }

	// rply2, err := c.POST_List(ctx, RequestPOST)
	// if err != nil {
	// 	log.Println("something went wrong", err)
	// }
	// log.Println(rply2)

}
