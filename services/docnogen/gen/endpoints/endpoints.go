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
	GetNextDocNoEndpoint endpoint.Endpoint

	ConsumeDocNoEndpoint endpoint.Endpoint
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

		GetNextDocNoEndpoint: getnextdocnoEndpoint,

		ConsumeDocNoEndpoint: consumedocnoEndpoint,
	}
}
