package rabbitmq

import (
	"fmt"
	"time"

	"github.com/wagslane/go-rabbitmq"
)

func NewConnection(url string) (*rabbitmq.Conn, error) {
	conn, err := rabbitmq.NewConn(
		url,
		rabbitmq.WithConnectionOptionsReconnectInterval(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	return conn, nil
}
