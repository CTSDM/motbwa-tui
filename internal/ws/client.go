package ws

import (
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebsocketConnection interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	WriteControl(messageType int, data []byte, deadline time.Time) error
	SetReadDeadline(t time.Time) error
	NextReader() (messageType int, r io.Reader, err error)
	Close() error
}

type Room struct {
	id   uuid.UUID
	name string
}

type ClientManager struct {
	Conn        WebsocketConnection
	CurrentRoom *Room
	Rooms       []Room
	msgChan     chan Message
	egress      chan string
	user        userInfo
}

type userInfo struct {
	name string
	id   uuid.UUID
}

func NewClientManager(c WebsocketConnection, username string, userID uuid.UUID, r Room) *ClientManager {
	return &ClientManager{
		Conn:        c,
		CurrentRoom: &r,
		Rooms:       []Room{r},
		msgChan:     make(chan Message),
		egress:      make(chan string),
		user:        userInfo{name: username, id: userID},
	}
}

func (c *ClientManager) sendMessages() {
	for {
		writeMessage(c.Conn, c.user, c.CurrentRoom.id, <-c.egress)
	}
}

func (c *ClientManager) Close() {
	deadline := time.Now().Add(1 * time.Second)
	if err := c.Conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		deadline,
	); err != nil {
		if err := c.Conn.Close(); err != nil {
			log.Fatalf("Error while closing the websocket connection: %s", err)
		}
	}

	// Set deadline for reading the next message
	if err := c.Conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		if err := c.Conn.Close(); err != nil {
			log.Fatalf("Error while closing the websocket connection: %s", err)
		}
	}
	// Read messages until the close message is confirmed
	for {
		_, _, err := c.Conn.NextReader()
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			break
		}
		if err != nil {
			break
		}
	}
}

func (c *ClientManager) readMessages() {
	for {
		var event Event
		_, payload, err := c.Conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Println("normal closure of the websocket connection...")
			return
		} else if err != nil {
			log.Fatal(err)
		}

		if err := json.Unmarshal(payload, &event); err != nil {
			log.Printf("error unmarshaling event: %v", err)
			return
		}
		// for now there are no rooms
		// we assign the incoming event to the message object
		message := Message{Content: string(event.Message.Content), Sender: event.Message.Sender}
		c.msgChan <- message
	}
}

func writeMessage(c WebsocketConnection, user userInfo, room uuid.UUID, message string) {
	payload := getMessageToSend(room, user, message)
	err := c.WriteMessage(websocket.TextMessage, payload)
	if err != nil {
		log.Println("something went wrong while writing the message!!!!")
		log.Fatal(err)
	}
}

func (c *ClientManager) MessageChannel() <-chan Message {
	return c.msgChan
}

func (c *ClientManager) SetEgress(msg string) {
	c.egress <- msg
}
