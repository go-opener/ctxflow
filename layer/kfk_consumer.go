package layer

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"github.com/go-opener/ctxflow/v2/puzzle"
)

type IConsumer interface {
	IFlow
    Run(args []string) error
    Setup(sarama.ConsumerGroupSession) error
    Cleanup(sarama.ConsumerGroupSession) error
    ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error
}

type ConsumeHandler func(sarama.ConsumerGroupSession, *sarama.ConsumerMessage) error

type KFKConsumer struct {
	Flow
    handler ConsumeHandler
    Ready   chan bool
	Topics []string
}

func (entity *KFKConsumer) LoopKafka(handler ConsumeHandler) {
	var err error

	client := puzzle.GetKafkaClient()

	ctx, cancel := context.WithCancel(entity.GetContext())

	entity.handler = handler

	//fmt.Printf("puzzle.GetKafkaBrokers():%+v",puzzle.GetKafkaBrokers())

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx,entity.Topics, entity); err != nil {
				entity.LogAndExit("Error from consumer: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
			//entity.Ready = make(chan bool)
		}
	}()
	wg.Wait()
	//<-entity.Ready // Await till the consumer has been set up
	entity.LogInfof("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-entity.GetContext().Done():
		entity.LogInfof("terminating: context cancelled")
	case <-sigterm:
		entity.LogInfof("terminating: via signal")
	}
	cancel()

	if err = client.Close(); err != nil {
		entity.LogAndExit("Error closing client: %v", err)
	}

}

func (entity *KFKConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(entity.Ready)
	return nil
}

func (entity *KFKConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (entity *KFKConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		err := entity.handler(session, message)
		if err != nil {
			entity.LogErrorf("Error handle consumer message: %v", err)
		}
	}
	return nil
}

func (entity *KFKConsumer) LogAndExit(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	entity.LogError(s)
	panic(s)
}