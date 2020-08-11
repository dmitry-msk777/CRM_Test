package protomodul

import (
	//pb "../CRM_Test/proto"
	"context"
	"net"

	enginecrm "github.com/dmitry-msk777/CRM_Test/enginecrm"
	pb "github.com/dmitry-msk777/CRM_Test/proto"
	rootsctuct "github.com/dmitry-msk777/CRM_Test/rootdescription"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

//Protobuff
type server struct{}

func (s *server) GET_List(ctx context.Context, in *pb.RequestGET) (*pb.ResponseGET, error) {

	id := in.CustomerId

	Customer_struct_out, err := enginecrm.EngineCRMv.FindOneRow(enginecrm.EngineCRMv.DataBaseType, id, rootsctuct.Global_settingsV)

	if err != nil {
		enginecrm.EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		return nil, nil
	}

	response := &pb.ResponseGET{
		CustomerId:    Customer_struct_out.Customer_id,
		CustomerName:  Customer_struct_out.Customer_name,
		CustomerType:  Customer_struct_out.Customer_type,
		CustomerEmail: Customer_struct_out.Customer_email,
	}

	return response, nil
}

func (s *server) POST_List(ctx context.Context, in *pb.RequestPOST) (*pb.ResponsePOST, error) {

	Customer_struct_out := rootsctuct.Customer_struct{
		Customer_id:    in.CustomerId,
		Customer_name:  in.CustomerName,
		Customer_type:  in.CustomerType,
		Customer_email: in.CustomerEmail,
	}

	err := enginecrm.EngineCRMv.AddChangeOneRow(enginecrm.EngineCRMv.DataBaseType, Customer_struct_out, rootsctuct.Global_settingsV)

	if err != nil {
		enginecrm.EngineCRMv.LoggerCRM.ErrorLogger.Println(err.Error())
		return nil, err
	}

	return &pb.ResponsePOST{CustomerId: "True"}, nil
}

//Protobuff end

func InitgRPC() {
	listener, err := net.Listen("tcp", ":5300")

	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterCRMswapServer(grpcServer, &server{})
	grpcServer.Serve(listener)
}
