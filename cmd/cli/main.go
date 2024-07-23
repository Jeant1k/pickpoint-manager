package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/cli"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/config"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/kafka"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/module"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/outbox"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/repository"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/repository/transactor"
)

func main() {
	fmt.Println("\tПривет! Эта программа по управлению пунктом выдачи.")

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

	transactionManager := transactor.NewTransactionManager(pool)
	repo := repository.NewRepository(transactionManager)
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
	}

	commands := cli.NewCLI(cli.Deps{
		Module:      pickPointManager,
		PickPointId: pickPointManager.PickPointId,
		Outbox:      ob,
		Producer:    producer,
		Consumer:    consumer,
	})

	taskQueue := make(chan cli.Task, 10)
	var wg sync.WaitGroup

	workerCount := viper.GetInt("workerCount")
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ctx, taskQueue, &wg)
	}

	go func() {
		if err := commands.Run(ctx, taskQueue); err != nil {
			fmt.Println(err)
			cancel()
		}
		close(taskQueue)
	}()

	wg.Wait()

	fmt.Println("\tДо скорых встреч!")
}

func worker(ctx context.Context, taskQueue chan cli.Task, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case task, ok := <-taskQueue:
			if !ok {
				return
			}
			task.Execute()
		case <-ctx.Done():
			return
		}
	}
}

func handleSignals(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	fmt.Println("\n\tЗавершаю работу по сигналу")
	cancel()
}
