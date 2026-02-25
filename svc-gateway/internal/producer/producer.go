package producer

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

// Producer wraps a kafka-go Writer and provides typed protobuf publishing.
type Producer struct {
	writer *kafka.Writer
}

// New returns a Producer connected to the given broker and topic.
func New(brokers []string, topic string) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
	}
	return &Producer{writer: w}
}

// Publish serializes msg as protobuf and sends it to Kafka.
func (p *Producer) Publish(ctx context.Context, key string, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("producer: marshal: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: data,
	})
}

// Close flushes and closes the underlying writer.
func (p *Producer) Close() error {
	return p.writer.Close()
}
