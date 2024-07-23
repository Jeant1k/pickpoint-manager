package kafka

import (
	"context"
	"log"

	"github.com/IBM/sarama"
	"gitlab.ozon.dev/evgeniy_tikh0n08/homework-1/internal/color"
)

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
}

func NewConsumer(brokers []string, topic string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, "consumer-group", config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		topic:         topic,
	}, nil
}

func (c *Consumer) StartConsumer(ctx context.Context) error {
	handler := &consumerGroupHandler{}

	for {
		if err := c.consumerGroup.Consume(ctx, []string{c.topic}, handler); err != nil {
			log.Printf("Произошла ошибка consumer: %v", err)
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *Consumer) Close() error {
	return c.consumerGroup.Close()
}

type consumerGroupHandler struct{}

func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		color.PrintYellow("Лог: значение = " + string(msg.Value) + " topic = " + msg.Topic)
		sess.MarkMessage(msg, "")
	}
	return nil
}
