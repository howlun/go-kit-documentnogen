package docnogen_endpoints

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/ratelimit"
	pb "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/pb"
	"github.com/sony/gobreaker"
	context "golang.org/x/net/context"
	"golang.org/x/time/rate"
)

//var _ = endpoint.Chain
var _ = fmt.Errorf

//var _ = context.Background

type StreamEndpoint func(server interface{}, req interface{}) (err error)

type Endpoints struct {
	GenerateBulkDocNoFormatEndpoint endpoint.Endpoint

	GenerateDocNoFormatEndpoint endpoint.Endpoint

	GetNextDocNoEndpoint endpoint.Endpoint

	ConsumeDocNoEndpoint endpoint.Endpoint
}

func (e *Endpoints) GenerateBulkDocNoFormat(ctx context.Context, in *pb.GenerateBulkDocNoFormatRequest) (*pb.GenerateBulkDocNoFormatResponse, error) {
	out, err := e.GenerateBulkDocNoFormatEndpoint(ctx, in)
	if err != nil {
		return &pb.GenerateBulkDocNoFormatResponse{}, err
	}
	return out.(*pb.GenerateBulkDocNoFormatResponse), err
}

func (e *Endpoints) GenerateDocNoFormat(ctx context.Context, in *pb.GenerateDocNoFormatRequest) (*pb.GenerateDocNoFormatResponse, error) {
	out, err := e.GenerateDocNoFormatEndpoint(ctx, in)
	if err != nil {
		return &pb.GenerateDocNoFormatResponse{}, err
	}
	return out.(*pb.GenerateDocNoFormatResponse), err
}

func (e *Endpoints) GetNextDocNo(ctx context.Context, in *pb.GetNextDocNoRequest) (*pb.GetNextDocNoResponse, error) {
	out, err := e.GetNextDocNoEndpoint(ctx, in)
	if err != nil {
		return &pb.GetNextDocNoResponse{}, err
	}
	return out.(*pb.GetNextDocNoResponse), err
}

func (e *Endpoints) ConsumeDocNo(ctx context.Context, in *pb.ConsumeDocNoRequest) (*pb.ConsumeDocNoResponse, error) {
	out, err := e.ConsumeDocNoEndpoint(ctx, in)
	if err != nil {
		return &pb.ConsumeDocNoResponse{}, err
	}
	return out.(*pb.ConsumeDocNoResponse), err
}

func MakeGenerateBulkDocNoFormatEndpoint(svc pb.DocNoGenServiceServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*pb.GenerateBulkDocNoFormatRequest)
		rep, err := svc.GenerateBulkDocNoFormat(ctx, req)
		if err != nil {
			return &pb.GenerateBulkDocNoFormatResponse{}, err
		}
		return rep, nil
	}
}

func MakeGenerateDocNoFormatEndpoint(svc pb.DocNoGenServiceServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*pb.GenerateDocNoFormatRequest)
		rep, err := svc.GenerateDocNoFormat(ctx, req)
		if err != nil {
			return &pb.GenerateDocNoFormatResponse{}, err
		}
		return rep, nil
	}
}

func MakeGetNextDocNoEndpoint(svc pb.DocNoGenServiceServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*pb.GetNextDocNoRequest)
		rep, err := svc.GetNextDocNo(ctx, req)
		if err != nil {
			return &pb.GetNextDocNoResponse{}, err
		}
		return rep, nil
	}
}

func MakeConsumeDocNoEndpoint(svc pb.DocNoGenServiceServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*pb.ConsumeDocNoRequest)
		rep, err := svc.ConsumeDocNo(ctx, req)
		if err != nil {
			return &pb.ConsumeDocNoResponse{}, err
		}
		return rep, nil
	}
}

func MakeEndpoints(svc pb.DocNoGenServiceServer, logger log.Logger, duration metrics.Histogram) Endpoints {

	var generateBulkDocNoFormatEndpoint endpoint.Endpoint
	{
		generateBulkDocNoFormatEndpoint = MakeGenerateBulkDocNoFormatEndpoint(svc)
		generateBulkDocNoFormatEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 10))(generateBulkDocNoFormatEndpoint)
		generateBulkDocNoFormatEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(generateBulkDocNoFormatEndpoint)
		generateBulkDocNoFormatEndpoint = LoggingMiddleware(log.With(logger, "method", "GenerateBulkDocNoFormat"))(generateBulkDocNoFormatEndpoint)
		generateBulkDocNoFormatEndpoint = InstrumentingMiddleware(duration.With("method", "GenerateBulkDocNoFormat"))(generateBulkDocNoFormatEndpoint)
	}

	var generatedocnoformatEndpoint endpoint.Endpoint
	{
		generatedocnoformatEndpoint = MakeGenerateDocNoFormatEndpoint(svc)
		generatedocnoformatEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 10))(generatedocnoformatEndpoint)
		generatedocnoformatEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(generatedocnoformatEndpoint)
		generatedocnoformatEndpoint = LoggingMiddleware(log.With(logger, "method", "GenerateDocNoFormat"))(generatedocnoformatEndpoint)
		generatedocnoformatEndpoint = InstrumentingMiddleware(duration.With("method", "GenerateDocNoFormat"))(generatedocnoformatEndpoint)
	}

	var getnextdocnoEndpoint endpoint.Endpoint
	{
		getnextdocnoEndpoint = MakeGetNextDocNoEndpoint(svc)
		getnextdocnoEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 10))(getnextdocnoEndpoint)
		getnextdocnoEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getnextdocnoEndpoint)
		getnextdocnoEndpoint = LoggingMiddleware(log.With(logger, "method", "GetNextDocNo"))(getnextdocnoEndpoint)
		getnextdocnoEndpoint = InstrumentingMiddleware(duration.With("method", "GetNextDocNo"))(getnextdocnoEndpoint)
	}

	var consumedocnoEndpoint endpoint.Endpoint
	{
		consumedocnoEndpoint = MakeConsumeDocNoEndpoint(svc)
		consumedocnoEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 10))(consumedocnoEndpoint)
		consumedocnoEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(consumedocnoEndpoint)
		consumedocnoEndpoint = LoggingMiddleware(log.With(logger, "method", "ConsumeDocNo"))(consumedocnoEndpoint)
		consumedocnoEndpoint = InstrumentingMiddleware(duration.With("method", "ConsumeDocNo"))(consumedocnoEndpoint)
	}

	return Endpoints{

		GenerateBulkDocNoFormatEndpoint: generateBulkDocNoFormatEndpoint,

		GenerateDocNoFormatEndpoint: generatedocnoformatEndpoint,

		GetNextDocNoEndpoint: getnextdocnoEndpoint,

		ConsumeDocNoEndpoint: consumedocnoEndpoint,
	}
}
