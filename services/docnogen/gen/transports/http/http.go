package docnogen_httptransport

import (
	"encoding/json"
	"errors"
	stdLog "log"
	"net/http"

	context "golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	endpoints "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/endpoints"
	pb "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/pb"
)

//var _ = log.Printf
//var _ = gokit_endpoint.Chain
//var _ = httptransport.NewClient

func MakeGenerateBulkDocNoFormatHandler(_ context.Context, svc pb.DocNoGenServiceServer, endpoint endpoint.Endpoint, logger log.Logger) *httptransport.Server {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorLogger(logger),
	}

	return httptransport.NewServer(
		endpoint,
		decodeGenerateBulkDocNoFormatRequest,
		encodeGenerateBulkDocNoFormatResponse,
		options...,
	)
}

func decodeGenerateBulkDocNoFormatRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req pb.GenerateBulkDocNoFormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func encodeGenerateBulkDocNoFormatResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func MakeGenerateDocNoFormatHandler(_ context.Context, svc pb.DocNoGenServiceServer, endpoint endpoint.Endpoint, logger log.Logger) *httptransport.Server {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorLogger(logger),
	}

	return httptransport.NewServer(
		endpoint,
		decodeGenerateDocNoFormatRequest,
		encodeGenerateDocNoFormatResponse,
		options...,
	)
}

func decodeGenerateDocNoFormatRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req pb.GenerateDocNoFormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func encodeGenerateDocNoFormatResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func MakeGetNextDocNoHandler(_ context.Context, svc pb.DocNoGenServiceServer, endpoint endpoint.Endpoint, logger log.Logger) *httptransport.Server {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorLogger(logger),
	}

	return httptransport.NewServer(
		endpoint,
		decodeGetNextDocNoRequest,
		encodeGetNextDocNoResponse,
		options...,
	)
}

func decodeGetNextDocNoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req pb.GetNextDocNoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func encodeGetNextDocNoResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func MakeConsumeDocNoHandler(_ context.Context, svc pb.DocNoGenServiceServer, endpoint endpoint.Endpoint, logger log.Logger) *httptransport.Server {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorLogger(logger),
	}

	return httptransport.NewServer(
		endpoint,
		decodeConsumeDocNoRequest,
		encodeConsumeDocNoResponse,
		options...,
	)
}

func decodeConsumeDocNoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req pb.ConsumeDocNoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func encodeConsumeDocNoResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func RegisterHandlers(ctx context.Context, svc pb.DocNoGenServiceServer, mux *http.ServeMux, endpoints endpoints.Endpoints, logger log.Logger) error {

	stdLog.Println("new HTTP endpoint: \"/GenerateBulkDocNoFormat\" (service=Docnogen)")
	mux.Handle("/GenerateBulkDocNoFormat", MakeGenerateBulkDocNoFormatHandler(ctx, svc, endpoints.GenerateBulkDocNoFormatEndpoint, logger))

	stdLog.Println("new HTTP endpoint: \"/GenerateDocNoFormat\" (service=Docnogen)")
	mux.Handle("/GenerateDocNoFormat", MakeGenerateDocNoFormatHandler(ctx, svc, endpoints.GenerateDocNoFormatEndpoint, logger))

	stdLog.Println("new HTTP endpoint: \"/GetNextDocNo\" (service=Docnogen)")
	mux.Handle("/GetNextDocNo", MakeGetNextDocNoHandler(ctx, svc, endpoints.GetNextDocNoEndpoint, logger))

	stdLog.Println("new HTTP endpoint: \"/ConsumeDocNo\" (service=Docnogen)")
	mux.Handle("/ConsumeDocNo", MakeConsumeDocNoHandler(ctx, svc, endpoints.ConsumeDocNoEndpoint, logger))

	return nil
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

func err2code(err error) int {
	/*
		switch err {
		case addservice.ErrTwoZeroes, addservice.ErrMaxSizeExceeded, addservice.ErrIntOverflow:
			return http.StatusBadRequest
		}
	*/
	return http.StatusInternalServerError
}

func errorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

type errorWrapper struct {
	Error string `json:"error"`
}
