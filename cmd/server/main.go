package main

import (
	"context"
	"fmt"
	stdLog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/handlers"
	"github.com/urfave/cli"
	"google.golang.org/grpc"

	docnogensvc "github.com/howlun/go-kit-documentnogen/services/docnogen"
	docnogenendpoints "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/endpoints"
	docnogenpb "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/pb"
	docnogengrpctransport "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/transports/grpc"
	docnogenhttptransport "github.com/howlun/go-kit-documentnogen/services/docnogen/gen/transports/http"

	docnogenmodel "github.com/howlun/go-kit-documentnogen/services/docnogen/models"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	app := cli.NewApp()
	app.Name = "docnogen-server"
	app.Usage = "Document Number Generator Server (gRPC + http)"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Par How Lun",
			Email: "howlun.par@gmail.com",
		},
	}
	app.Copyright = fmt.Sprintf("(c) %d Par How Lun", time.Now().Year())
	app.Flags = []cli.Flag{
		/*
			cli.StringSliceFlag{
				Name:   "brokers, b",
				Usage:  "List of Kafka Brokers",
				EnvVar: "KAFKA_BROKERS",
				Value:  &cli.StringSlice{"127.0.0.1:9092"},
			},
		*/
		cli.StringFlag{
			Name:  "httpaddr",
			Value: ":12000",
			Usage: "Http Server Address",
		},
		cli.StringFlag{
			Name:  "grpcaddr",
			Value: ":13000",
			Usage: "GRPC Server Address",
		},
		cli.StringFlag{
			Name:  "mongoaddr",
			Value: "localhost:27017",
			Usage: "Mongo DB Server Address",
		},
		cli.StringFlag{
			Name:  "mongodbname",
			Value: "docnogen_v1",
			Usage: "Mongo DB Name",
		},
		cli.StringFlag{
			Name:  "mongoauthusername",
			Value: "",
			Usage: "Mongo DB Auth Username",
		},
		cli.StringFlag{
			Name:  "mongoauthpassword",
			Value: "",
			Usage: "Mongo DB Auth Password",
		},
		cli.StringFlag{
			Name:  "httplog",
			Value: "log/http.log",
			Usage: "HTTP log directory and filename",
		},
	}
	app.Action = runMain
	err := app.Run(os.Args)
	if err != nil {
		stdLog.Fatal(err)
	}
}

func runMain(c *cli.Context) error {
	// show Help info whenever server starts
	cli.ShowAppHelp(c)

	mux := http.NewServeMux()
	ctx := context.Background()
	errc := make(chan error)
	s := grpc.NewServer()
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	/*
		var kafkaSyncProducer sarama.SyncProducer
		{
			config := sarama.NewConfig()
			config.Producer.RequiredAcks = sarama.WaitForAll
			config.Producer.Retry.Max = 5
			config.Producer.Return.Successes = true
			var err error
			kafkaSyncProducer, err = sarama.NewSyncProducer(c.StringSlice("brokers"), config)
			if err != nil {
				stdlog.Printf("Failed to initiate sarama.SyncProducer: %v", err)
				os.Exit(-1)
			}
		}
	*/
	var duration metrics.Histogram
	{
		// Endpoint-level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "howlun",
			Subsystem: "docnogen",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}
	mux.Handle("/metrics", promhttp.Handler())

	{
		dbclient := docnogenmodel.NewDBClient(c.String("mongoaddr"), c.String("mongodbname"), c.String("mongoauthusername"), c.String("mongoauthpassword"))
		docNoRepo := docnogenmodel.NewDocNoRepository(dbclient)

		docNoFormatterSvc := docnogensvc.NewDocnoformatterService()
		svc := docnogensvc.NewDocnogenService(docNoRepo, docNoFormatterSvc)
		endpoints := docnogenendpoints.MakeEndpoints(svc, logger, duration)
		srv := docnogengrpctransport.MakeGRPCServer(ctx, endpoints, logger)
		docnogenpb.RegisterDocNoGenServiceServer(s, srv)
		docnogenhttptransport.RegisterHandlers(ctx, svc, mux, endpoints, logger)
	}

	// start servers
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger := log.With(logger, "transport", "HTTP")
		logger.Log("addr", c.String("httpaddr"))

		// http log writer
		err := ensureDir(c.String("httplog"))
		if err != nil {
			errc <- err
		}
		httpLogFile, err := os.OpenFile(c.String("httplog"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			errc <- err
		}

		// gorilla/handlers LoggingHandler is used for logging HTTP requests in the Apache Common Log Format
		errc <- http.ListenAndServe(c.String("httpaddr"), handlers.LoggingHandler(httpLogFile, mux))
	}()

	go func() {
		logger := log.With(logger, "transport", "gRPC")
		ln, err := net.Listen("tcp", c.String("grpcaddr"))
		if err != nil {
			errc <- err
			return
		}
		logger.Log("addr", c.String("grpcaddr"))
		errc <- s.Serve(ln)
	}()

	logger.Log("exit", <-errc)
	return nil
}

func ensureDir(fileName string) error {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			return merr
		}
	}

	return nil
}
