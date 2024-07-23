package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	service "gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/api"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/cache"
	memorycache "gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/cache/memory"
	rediscache "gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/cache/redis"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/config"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/kafka"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/module"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/outbox"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/repository"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/repository/transactor"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/pkg/api/proto/pickpoint/v1/pickpoint/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	grpcPort    = ":50051"
	httpPort    = ":63342"
	metricsPort = ":9100"
)

var (
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:50051", "gRPC server endpoint")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handleSignals(cancel)

	config.Init()

	dbURL := viper.GetString("dbURL")
	pool, errConn := pgxpool.Connect(ctx, dbURL)
	if errConn != nil {
		fmt.Printf("Ошибка подключения к базе данных: %v\n", errConn)
		return
	}
	defer pool.Close()

	var cache cache.Cache
	if viper.GetBool("useRedis") {
		redisURL := viper.GetString("redis.url")
		redisPassword := viper.GetString("redis.password")
		redisTTL := viper.GetDuration("redis.ttl") * time.Second
		cache = rediscache.New(ctx, redisURL, redisPassword, redisTTL)
	} else {
		cache = memorycache.New(viper.GetDuration("memorycache.ttl") * time.Second)
	}

	transactionManager := transactor.NewTransactionManager(pool)
	repo := repository.NewRepository(transactionManager, cache)
	pickPointManager := module.NewModule(module.Deps{
		Repository:         repo,
		TransactionMaganer: transactionManager,
	})

	var producer *kafka.Producer
	var consumer *kafka.Consumer
	var ob *outbox.Outbox
	if viper.GetBool("useKafka") {
		brokers := viper.GetStringSlice("kafka.brokers")
		topic := viper.GetString("kafka.topic")

		p, errConn := kafka.NewProducer(brokers, topic)
		if errConn != nil {
			fmt.Printf("Ошибка подключения к Kafka: %v\n", errConn)
			return
		}
		defer p.Close()
		producer = p

		c, errConn := kafka.NewConsumer(brokers, topic)
		if errConn != nil {
			fmt.Printf("Ошибка подключения к Kafka: %v\n", errConn)
			return
		}
		defer c.Close()
		consumer = c

		ob = outbox.NewOutbox(pool, producer)

		go ob.StartBackgroundProcessing(ctx)
		go c.StartConsumer(ctx)
	}

	pickPointService := service.NewPickPointService(service.Deps{
		Module:      pickPointManager,
		PickPointId: pickPointManager.PickPointId,
		Outbox:      ob,
		Producer:    producer,
		Consumer:    consumer,
		Ctx:         ctx,
	})

	lis, err := net.Listen("tcp", grpcPort)
	log.Println("Start")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pickpoint.RegisterPickpointServer(grpcServer, &pickPointService)

	reflection.Register(grpcServer)

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	errGw := pickpoint.RegisterPickpointHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if errGw != nil {
		log.Fatalf("failed to RegisterPickpointHandlerFromEndpoint: %v", errGw)
		return
	}

	go func() {
		gwServer := &http.Server{
			Addr: httpPort,
		}

		errHttp := gwServer.ListenAndServe()
		if errHttp != nil {
			log.Println(errHttp)
			return
		}

	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(metricsPort, nil)
		if err != nil {
			log.Fatalf("failed to start metrics server: %v", err)
		}
	}()

	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		cancel()
	}
}

func handleSignals(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	fmt.Println("\n\tЗавершаю работу по сигналу")
	cancel()
}
