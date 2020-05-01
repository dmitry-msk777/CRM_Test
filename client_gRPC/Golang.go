package main

import (
	"context"
	"log"
	"time"

	pb "../proto"
	"google.golang.org/grpc"
)

func main() {

	//Открываем соединение, grpc.WithInsecure() означает,
	//что шифрование не используется
	conn, err := grpc.Dial("localhost:5300", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	/*

	   Создаём нового клиента, используя соединение conn
	   Обратим внимание на название клиента и на название сервиса,
	   которое мы определили в proto-файле:

	   service Mailer {
	   rpc SendPass(MsgRequest) returns (MsgReply) {}
	   rpc RetrievePass(MsgRequest) returns (MsgReply) {}
	   }

	*/

	c := pb.NewCRMswapClient(conn)

	//Определяем контекст с таймаутом в 1 секунду
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	/*
	   Шлём запрос 1, ожидаем получение true в структуру rply
	   типа MsgReply, определённую в прото-файле как:

	   message MsgReply {
	   bool sent = 1;
	   }

	*/

	// //Retrive
	// RequestGET := &pb.RequestGET{
	// 	CustomerId: "777666",
	// }

	// rply, err := c.GET_List(ctx, RequestGET)
	// if err != nil {
	// 	log.Println("something went wrong", err)
	// }
	// log.Println(rply)

	// Record
	RequestPOST := &pb.RequestPOST{
		CustomerId:    "123007",
		CustomerName:  "Alex gRPC",
		CustomerType:  "Google",
		CustomerEmail: "gRPC@mail.ru",
	}

	rply2, err := c.POST_List(ctx, RequestPOST)
	if err != nil {
		log.Println("something went wrong", err)
	}
	log.Println(rply2)

}
