package ws

import "time"

type Message struct {
	Sender  string
	Content string
	Date    time.Time
}
