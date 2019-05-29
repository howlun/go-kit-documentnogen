package docnogensvc

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	pb "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/pb"
	context "golang.org/x/net/context"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(DocnogenService) DocnogenService

// LoggingMiddleware takes a logger as a dependency
// and returns a service middleware.
type loggingMiddleware struct {
	next   DocnogenService
	logger log.Logger
}

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next DocnogenService) DocnogenService {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

func (mw loggingMiddleware) GenerateBulkDocNoFormat(ctx context.Context, in *pb.GenerateBulkDocNoFormatRequest) (out *pb.GenerateBulkDocNoFormatResponse, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GenerateBulkDocNoFormat", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GenerateBulkDocNoFormat(ctx, in)
}

func (mw loggingMiddleware) GenerateDocNoFormat(ctx context.Context, in *pb.GenerateDocNoFormatRequest) (out *pb.GenerateDocNoFormatResponse, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GenerateDocNoFormat", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GenerateDocNoFormat(ctx, in)
}

func (mw loggingMiddleware) GetNextDocNo(ctx context.Context, in *pb.GetNextDocNoRequest) (out *pb.GetNextDocNoResponse, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetNextDocNo", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GetNextDocNo(ctx, in)
}

func (mw loggingMiddleware) ConsumeDocNo(ctx context.Context, in *pb.ConsumeDocNoRequest) (out *pb.ConsumeDocNoResponse, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "ConsumeDocNo", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.ConsumeDocNo(ctx, in)
}

// InstrumentingMiddleware returns a service middleware that instruments
// the number of integers summed and characters concatenated over the lifetime of
// the service.
type instrumentingMiddleware struct {
	ints  metrics.Counter
	chars metrics.Counter
	next  DocnogenService
}

func InstrumentingMiddleware(ints, chars metrics.Counter) Middleware {
	return func(next DocnogenService) DocnogenService {
		return instrumentingMiddleware{
			//ints:  ints,
			//chars: chars,
			next: next,
		}
	}
}

func (mw instrumentingMiddleware) GenerateBulkDocNoFormat(ctx context.Context, in *pb.GenerateBulkDocNoFormatRequest) (out *pb.GenerateBulkDocNoFormatResponse, err error) {
	v, err := mw.next.GenerateBulkDocNoFormat(ctx, in)
	// TODO: implement instrumenting logic here

	return v, err
}

func (mw instrumentingMiddleware) GenerateDocNoFormat(ctx context.Context, in *pb.GenerateDocNoFormatRequest) (out *pb.GenerateDocNoFormatResponse, err error) {
	v, err := mw.next.GenerateDocNoFormat(ctx, in)
	// TODO: implement instrumenting logic here

	return v, err
}

func (mw instrumentingMiddleware) GetNextDocNo(ctx context.Context, in *pb.GetNextDocNoRequest) (out *pb.GetNextDocNoResponse, err error) {
	v, err := mw.next.GetNextDocNo(ctx, in)
	// TODO: implement instrumenting logic here

	return v, err
}

func (mw instrumentingMiddleware) ConsumeDocNo(ctx context.Context, in *pb.ConsumeDocNoRequest) (out *pb.ConsumeDocNoResponse, err error) {
	v, err := mw.next.ConsumeDocNo(ctx, in)
	// TODO: implement instrumenting logic here

	return v, err
}
