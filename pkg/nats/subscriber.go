package nats

import (
	"fmt"
	"link/pkg/common"

	"github.com/nats-io/nats.go"
)

// ! NATS subscriber 패키지
type NatsSubscriber struct {
	conn *nats.Conn
}

func NewSubscriber(conn *nats.Conn) *NatsSubscriber {
	return &NatsSubscriber{conn: conn}
}

func (s *NatsSubscriber) SubscribeEvent(subject string, handler func(msg *nats.Msg)) error {
	_, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		fmt.Printf("NATS 이벤트 수신[TOPIC: %s]: %v ", subject, msg.Data)
		handler(msg)
	})

	if err != nil {
		fmt.Printf("NATS 이벤트 수신 오류[TOPIC: %s]: %v ", subject, err)
		return common.NewError(500, "NATS 이벤트 수신 오류", err)
	}
	fmt.Printf("NATS 이벤트 수신 성공[TOPIC: %s]", subject)
	return nil
}
