package puzzle

import "github.com/Shopify/sarama"

type IKafka sarama.ConsumerGroup

var KafkaClient IKafka

//deprecated
func SetKafkaClient(client IKafka) {
    KafkaClient = client
}

func GetKafkaClient() IKafka {
    return KafkaClient
}