package docnogen_clientgrpc

import (
	context "golang.org/x/net/context"

	jwt "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"

	endpoints "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/endpoints"
	pb "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/pb"
)

func New(conn *grpc.ClientConn, logger log.Logger) pb.DocNoGenServiceServer {

	var generatedocnoformatEndpoint endpoint.Endpoint
	{
		generatedocnoformatEndpoint = grpctransport.NewClient(
			conn,
			"docnogen.DocnogenService",
			"GenerateDocNoFormat",
			EncodeGenerateDocNoFormatRequest,
			DecodeGenerateDocNoFormatResponse,
			pb.GenerateDocNoFormatResponse{},
			append([]grpctransport.ClientOption{}, grpctransport.ClientBefore(jwt.FromGRPCContext()))...,
		).Endpoint()
	}

	var getnextdocnoEndpoint endpoint.Endpoint
	{
		getnextdocnoEndpoint = grpctransport.NewClient(
			conn,
			"docnogen.DocnogenService",
			"GetNextDocNo",
			EncodeGetNextDocNoRequest,
			DecodeGetNextDocNoResponse,
			pb.GetNextDocNoResponse{},
			append([]grpctransport.ClientOption{}, grpctransport.ClientBefore(jwt.FromGRPCContext()))...,
		).Endpoint()
	}

	var consumedocnoEndpoint endpoint.Endpoint
	{
		consumedocnoEndpoint = grpctransport.NewClient(
			conn,
			"docnogen.DocnogenService",
			"ConsumeDocNo",
			EncodeConsumeDocNoRequest,
			DecodeConsumeDocNoResponse,
			pb.ConsumeDocNoResponse{},
			append([]grpctransport.ClientOption{}, grpctransport.ClientBefore(jwt.FromGRPCContext()))...,
		).Endpoint()
	}

	return &endpoints.Endpoints{

		GenerateDocNoFormatEndpoint: generatedocnoformatEndpoint,

		GetNextDocNoEndpoint: getnextdocnoEndpoint,

		ConsumeDocNoEndpoint: consumedocnoEndpoint,
	}
}

func EncodeGenerateDocNoFormatRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.GenerateDocNoFormatRequest)
	return req, nil
}

func DecodeGenerateDocNoFormatResponse(_ context.Context, grpcResponse interface{}) (interface{}, error) {
	response := grpcResponse.(*pb.GenerateDocNoFormatResponse)
	return response, nil
}

func EncodeGetNextDocNoRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.GetNextDocNoRequest)
	return req, nil
}

func DecodeGetNextDocNoResponse(_ context.Context, grpcResponse interface{}) (interface{}, error) {
	response := grpcResponse.(*pb.GetNextDocNoResponse)
	return response, nil
}

func EncodeConsumeDocNoRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(*pb.ConsumeDocNoRequest)
	return req, nil
}

func DecodeConsumeDocNoResponse(_ context.Context, grpcResponse interface{}) (interface{}, error) {
	response := grpcResponse.(*pb.ConsumeDocNoResponse)
	return response, nil
}
