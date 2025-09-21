package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/CTSDM/motbwa-tui/internal/api"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func CreateConnection(s api.State) (*ClientManager, error) {
	header := make(http.Header)
	s.AddAuthTokensToHeader(&header)

	conn, _, err := websocket.DefaultDialer.Dial(s.Server.WebsocketURL, header)
	if err != nil {
		log.Printf("couldnt perform the handshake on %s: %s", s.Server.WebsocketURL, err)
		return nil, err
	}

	client := NewClientManager(conn, s.User.Username, s.User.UserID, Room{id: uuid.New(), name: "default"})

	go client.readMessages()
	go client.sendMessages()

	return client, nil
}

func getMessageToSend(room uuid.UUID, user userInfo, msg string) []byte {
	event := Event{
		Type: "send_message",
		Room: room,
		Message: Message{
			Sender:  user.name,
			Date:    time.Now(),
			Content: msg,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Fatal("something went wrong while marshaling the event")
	}
	return data
}

type Event struct {
	Type    string    `json:"type"`
	Room    uuid.UUID `json:"room"`
	Message Message   `json:"message"`
}
