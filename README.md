# DOCNOGEN_BE
A Golang microservice that generate document number

> API Doc: https://documenter.getpostman.com/view/5502222/S1TR3z87

## Golang Installation
```
$ cd  /tmp
$ wget –c https://storage.googleapis.com/golang/go1.11.5.linux-amd64.tar.gz
$ sudo tar -C /usr/local -xvzf go1.11.5.linux-amd64.tar.gz
$ mkdir –p ~/go/bin
$ mkdir –p ~/go/src
$ mkdir –p ~/go/pkg
$ sudo nano ~/.profile
```
 Add in the following lines to the bottom of the file:
```
export PATH=$PATH:/usr/local/go/bin
export GOPATH="$HOME/go"
export GOBIN="$GOPATH/bin"
```
Save and exit

```
$ source ~/.profile
$ go version
$ go env
```

## Source Code Installation
```
$ cd /tmp
$ git clone https://tc-systems.visualstudio.com/MCR_v2_MY/_git/DOCNOGEN_BE
```
Type in Username and Password
```
$ cd DOCNOGEN_BE/deployment/{dev or staging or prod}
$ bash deploy.sh
$ sudo systemctl status docnogen-api
```

## Steps to deploy to different environment
1. create **deploy.sh** and **docnogen-api.service** for {env} environment under the folder **DOCNOGEN_BE/deployment/{env}**
2. configure the system to run with different options (Change **docnogen-api.service** file)
```
GLOBAL OPTIONS:
   --httpaddr value           Http Server Address (default: ":12000")
   --grpcaddr value           GRPC Server Address (default: ":13000")
   --mongoaddr value          Mongo DB Server Address (default: "localhost:27017")
   --mongodbname value        Mongo DB Name (default: "docnogen_v1")
   --mongoauthusername value  Mongo DB Auth Username
   --mongoauthpassword value  Mongo DB Auth Password
   --httplog value            HTTP log directory and filename (default: "log/http.log")
   --help, -h                 show help
   --version, -v              print the version
```
3. rebuild the source by
```
$ cd ~/go/src/github.com/howlun/DOCNOGEN_BE/cmd/server/
$ go build
$ sudo systemctl restart docnogen-api
```

## Steps to test CORS with Google Chrome Console:
Open up **Google Developer Tools**, and then switch to **Console** and type following:
```
fetch('http://localhost:12000/GetNextDocNo', {
  method: 'POST',
  body: JSON.stringify({
    docCode: "AP",
	orgCode: "MAT",
	path: "AP/PO/YGN-HQ/19",
	variableMap: { "DOCTYPE": "PO", "BRHCD": "YGN-HQ", "YEAR": "19" },
	customFormat: "{{PREFIX}}{{SEQNO}}"
  }),
  headers: {
    'Content-type': 'application/json; charset=UTF-8'
  }
})
.then(res => res.json())
.then(console.log)
```

## Steps to change API parameters, and regenerate proto file
1. go to **DOCNOGEN_BE/services/docnogen/docnogen.proto**, make changes or add new api interface to the file
2. bring up the terminal, and type following:
```
$ cd <Project Root Directory>/services/docnogen
$ protoc --go_out=plugins=grpc:./gen/pb/ docnogen.proto
```
3. a new file will be generated in **DOCNOGEN_BE/services/docnogen/gen/pb/docnogen.pb.go**
4. check the new file, and rename the file to **pb.go** to replace to old version

## Manual steps to add new API interface
1. go to **DOCNOGEN_BE/services/docnogen/docnogen.proto**, make changes or add new api interface to the file
    ```
    service DocNoGenService {
    	rpc GenerateBulkDocNoFormat(GenerateBulkDocNoFormatRequest) returns (GenerateBulkDocNoFormatResponse) {}
    
    	...
    }
    
    message GenerateBulkDocNoFormatRequest {
    	string docCode = 1;
    	string orgCode = 2;
    	string path = 3;
    	map<string, string> variableMap = 4;
    	uint32 bulkNumber = 5;
    }
    
    message GenerateBulkDocNoFormatResponse {
        bool ok = 1;
        int32 errorCode = 2;
        string errorMessage = 3;
    
        message Result {
            string docNoString = 1;
            uint32 nextSeqNo = 2;
            int64 recordTimestamp = 3;
        }
        repeated Result results = 4;
    }
    ```

2. bring up the terminal, and type following:
	```
	$ cd <Project Root Directory>/services/docnogen
	$ protoc --go_out=plugins=grpc:./gen/pb/ docnogen.proto
	```
3. a new file will be generated in **DOCNOGEN_BE/services/docnogen/gen/pb/docnogen.pb.go**
4. check the new file, and rename the file to **pb.go** to replace to old version
5. go to **DOCNOGEN_BE/services/docnogen/service_test.go** and add the following
	```
	func Test_GenerateBulkDocNoFormat(t *testing.T) {
		// TODO: Implement testin for method logics here

		Convey("This isn't yet implemented", nil)
	}
	```
6. go to **DOCNOGEN_BE/services/docnogen/service.go** and add the following:
	```
	type DocnogenService interface {
		GenerateBulkDocNoFormat(ctx context.Context, in *pb.GenerateBulkDocNoFormatRequest) (out *pb.GenerateBulkDocNoFormatResponse, err error)
		...
	}

	func (s *docnogenService) GenerateBulkDocNoFormat(ctx context.Context, in *pb.GenerateBulkDocNoFormatRequest) (out *pb.GenerateBulkDocNoFormatResponse, err error) {
		// TODO
		return nil, nil
	}

	```
7. go to **DOCNOGEN_BE/services/docnogen/middlewares.go** and add the following:
	```
	func (mw loggingMiddleware) GenerateBulkDocNoFormat(ctx context.Context, in *pb.GenerateBulkDocNoFormatRequest) (out *pb.GenerateBulkDocNoFormatResponse, err error) {
		defer func(begin time.Time) {
			mw.logger.Log("method", "GenerateBulkDocNoFormat", "took", time.Since(begin), "err", err)
		}(time.Now())
		return mw.next.GenerateBulkDocNoFormat(ctx, in)
	}

	...

	func (mw instrumentingMiddleware) GenerateBulkDocNoFormat(ctx context.Context, in *pb.GenerateBulkDocNoFormatRequest) (out *pb.GenerateBulkDocNoFormatResponse, err error) {
		v, err := mw.next.GenerateBulkDocNoFormat(ctx, in)
		// TODO: implement instrumenting logic here

		return v, err
	}
	```
8. go to **DOCNOGEN_BE/services/docnogen/gen/endpoints/endpoints.go** and add the following:
	```
	type Endpoints struct {
		GenerateBulkDocNoFormatEndpoint endpoint.Endpoint
		...
	}

	func (e *Endpoints) GenerateBulkDocNoFormat(ctx context.Context, in *pb.GenerateBulkDocNoFormatRequest) (*pb.GenerateBulkDocNoFormatResponse, error) {
		out, err := e.GenerateBulkDocNoFormatEndpoint(ctx, in)
		if err != nil {
			return &pb.GenerateBulkDocNoFormatResponse{}, err
		}
		return out.(*pb.GenerateBulkDocNoFormatResponse), err
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

	...

	func MakeEndpoints(svc pb.DocNoGenServiceServer, logger log.Logger, duration metrics.Histogram) Endpoints {
	   
		var generateBulkDocNoFormatEndpoint endpoint.Endpoint
		{
			generateBulkDocNoFormatEndpoint = MakeGenerateBulkDocNoFormatEndpoint(svc)
			generateBulkDocNoFormatEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 10))(generateBulkDocNoFormatEndpoint)
			generateBulkDocNoFormatEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(generateBulkDocNoFormatEndpoint)
			generateBulkDocNoFormatEndpoint = LoggingMiddleware(log.With(logger, "method", "GenerateBulkDocNoFormat"))(generateBulkDocNoFormatEndpoint)
			generateBulkDocNoFormatEndpoint = InstrumentingMiddleware(duration.With("method", "GenerateBulkDocNoFormat"))(generateBulkDocNoFormatEndpoint)
		}
	   
	   ...

	   return Endpoints{
         
		   GenerateBulkDocNoFormatEndpoint: generateBulkDocNoFormatEndpoint,

			...
		}
	}

	```
9. go to **DOCNOGEN_BE/services/docnogen/gen/transports/grpc/grpc.go** and add the following:
	```
	func MakeGRPCServer(_ context.Context, endpoints endpoints.Endpoints, logger log.Logger) pb.DocNoGenServiceServer {
		options := []grpctransport.ServerOption{
			grpctransport.ServerErrorLogger(logger),
		}

		return &grpcServer{

			generatebulkdocnoformat: grpctransport.NewServer(
				endpoints.GenerateBulkDocNoFormatEndpoint,
				decodeGenerateBulkDocNoFormatRequest,
				encodeGenerateBulkDocNoFormatResponse,
				options...,
			),

			...
		}
	}

	type grpcServer struct {
		generatebulkdocnoformat grpctransport.Handler

		...
	}

	func (s *grpcServer) GenerateBulkDocNoFormat(ctx context.Context, req *pb.GenerateBulkDocNoFormatRequest) (*pb.GenerateBulkDocNoFormatResponse, error) {
		_, rep, err := s.generatebulkdocnoformat.ServeGRPC(ctx, req)
		if err != nil {
			return nil, err
		}
		return rep.(*pb.GenerateBulkDocNoFormatResponse), nil
	}

	func decodeGenerateBulkDocNoFormatRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
		return grpcReq, nil
	}

	func encodeGenerateBulkDocNoFormatResponse(_ context.Context, response interface{}) (interface{}, error) {
		resp := response.(*pb.GenerateBulkDocNoFormatResponse)
		return resp, nil
	}
	```
10. go to **DOCNOGEN_BE/services/docnogen/gen/transports/http/http.go** and add the following:
	```
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

	...

	func RegisterHandlers(ctx context.Context, svc pb.DocNoGenServiceServer, mux *http.ServeMux, endpoints endpoints.Endpoints, logger log.Logger) error {
	   
		stdLog.Println("new HTTP endpoint: \"/GenerateBulkDocNoFormat\" (service=Docnogen)")
		mux.Handle("/GenerateBulkDocNoFormat", MakeGenerateBulkDocNoFormatHandler(ctx, svc, endpoints.GenerateBulkDocNoFormatEndpoint, logger))
	   
	   ...
	}
	```
11. go to **DOCNOGEN_BE/services/docnogen/gen/client/grpc/client.go** and add the following:
    ```
    func New(conn *grpc.ClientConn, logger log.Logger) pb.DocNoGenServiceServer {
		var generateBulkDocNoFormatEndpoint endpoint.Endpoint
		{
			generateBulkDocNoFormatEndpoint = grpctransport.NewClient(
				conn,
				"docnogen.DocnogenService",
				"GenerateBulkDocNoFormat",
				EncodeGenerateBulkDocNoFormatRequest,
				DecodeGenerateBulkDocNoFormatResponse,
				pb.GenerateBulkDocNoFormatResponse{},
				append([]grpctransport.ClientOption{}, grpctransport.ClientBefore(jwt.FromGRPCContext()))...,
			).Endpoint()
    	}
    
    	...
    
		return &endpoints.Endpoints{
			GenerateBulkDocNoFormatEndpoint: generateBulkDocNoFormatEndpoint,

			...
		}
	}

    func EncodeGenerateBulkDocNoFormatRequest(_ context.Context, request interface{}) (interface{}, error) {
    	req := request.(*pb.GenerateBulkDocNoFormatRequest)
    	return req, nil
    }
    
    func DecodeGenerateBulkDocNoFormatResponse(_ context.Context, grpcResponse interface{}) (interface{}, error) {
    	response := grpcResponse.(*pb.GenerateBulkDocNoFormatResponse)
    	return response, nil
    }
    ```

12. go to terminal to build the source to check for errors with these commands:
  ```
	$ cd <Project Root Directory>/services/docnogen/cmd/server
	$ go build
  ```
13. if no errors and build is successful, it means the codes that we put in through the steps above to wire up the new API generated are done. It is time to put in the logic in **DOCNOGEN_BE/services/docnogen/service.go**:
  ```
	func (s *docnogenService) GenerateBulkDocNoFormat(ctx context.Context, in *pb.GenerateBulkDocNoFormatRequest) (out *pb.GenerateBulkDocNoFormatResponse, err error) {
		// TODO
		return nil, nil
	}
  ```
