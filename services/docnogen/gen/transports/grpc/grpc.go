package docnogen_grpctransport

import (
	"fmt"

	"github.com/go-kit/kit/log"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	endpoints "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/endpoints"
	pb "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/pb"
	context "golang.org/x/net/context"
)

// avoid import errors
var _ = fmt.Errorf

func MakeGRPCServer(_ context.Context, endpoints endpoints.Endpoints, logger log.Logger) pb.DocNoGenServiceServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
	}

	return &grpcServer{

		generatedocnoformat: grpctransport.NewServer(
			endpoints.GenerateDocNoFormatEndpoint,
			decodeGenerateDocNoFormatRequest,
			encodeGenerateDocNoFormatResponse,
			options...,
		),

		getnextdocno: grpctransport.NewServer(
			endpoints.GetNextDocNoEndpoint,
			decodeGetNextDocNoRequest,
			encodeGetNextDocNoResponse,
			options...,
		),

		consumedocno: grpctransport.NewServer(
			endpoints.ConsumeDocNoEndpoint,
			decodeConsumeDocNoRequest,
			encodeConsumeDocNoResponse,
			options...,
		),
	}
}

type grpcServer struct {
	generatedocnoformat grpctransport.Handler

	getnextdocno grpctransport.Handler

	consumedocno grpctransport.Handler
}

func (s *grpcServer) GenerateDocNoFormat(ctx context.Context, req *pb.GenerateDocNoFormatRequest) (*pb.GenerateDocNoFormatResponse, error) {
	_, rep, err := s.generatedocnoformat.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.GenerateDocNoFormatResponse), nil
}

func decodeGenerateDocNoFormatRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	return grpcReq, nil
}

func encodeGenerateDocNoFormatResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.GenerateDocNoFormatResponse)
	return resp, nil
}

func (s *grpcServer) GetNextDocNo(ctx context.Context, req *pb.GetNextDocNoRequest) (*pb.GetNextDocNoResponse, error) {
	_, rep, err := s.getnextdocno.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.GetNextDocNoResponse), nil
}

func decodeGetNextDocNoRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	return grpcReq, nil
}

func encodeGetNextDocNoResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.GetNextDocNoResponse)
	return resp, nil
}

func (s *grpcServer) ConsumeDocNo(ctx context.Context, req *pb.ConsumeDocNoRequest) (*pb.ConsumeDocNoResponse, error) {
	_, rep, err := s.consumedocno.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.ConsumeDocNoResponse), nil
}

func decodeConsumeDocNoRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	return grpcReq, nil
}

func encodeConsumeDocNoResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*pb.ConsumeDocNoResponse)
	return resp, nil
}

type streamHandler interface {
	Do(server interface{}, req interface{}) (err error)
}

type server struct {
	e endpoints.StreamEndpoint
}

func (s server) Do(server interface{}, req interface{}) (err error) {
	if err := s.e(server, req); err != nil {
		return err
	}
	return nil
}
