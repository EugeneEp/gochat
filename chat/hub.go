package chat

import(
	"encoding/json"
	"os"
	"log"
)


// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *Client

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Create new channel
	channel chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]map[*Client]bool),
		channel: 	make(chan *Client),
	}
}

func (h *Hub) GetChannels() []byte{
	var channels []string
	for v := range h.clients{
		channels = append(channels, v)
	}
	js, err := json.Marshal(channels)
	if err != nil{
		log.Printf("error: %v", err)
	}
	return []byte(js)
} 

func (h *Hub) Run() {
	h.clients["MainChannel"] = make(map[*Client]bool)
	for {
		select {
		case client := <-h.register:
			chName := client.ChannelName
			if _, ok := h.clients[chName]; ok{
				h.clients[chName][client] = true
			}else{
				log.Println("Комната не найдена")
			}
		case client := <-h.unregister:
			chName := client.ChannelName
			if _, ok := h.clients[chName]; ok {
				if _, okClient := h.clients[chName][client]; okClient{
					delete(h.clients[chName], client)
					close(client.send)
				}
			}
		case sender := <-h.broadcast:
			chName := sender.ChannelName
			for client := range h.clients[chName] {
				select {
					case client.send <- sender.Message:
					default:
						close(client.send)
						delete(h.clients[sender.ChannelName], client)
				}
			}
		case sender := <-h.channel:
			msg := make(map[string]string)
			err := json.Unmarshal(sender.Message, &msg)
			if err != nil {
				log.Printf("error: %v", err)
			}
			if msg["action"] == "create"{
				if _, ok := h.clients[msg["channel"]]; ok{
					sender.Message = []byte(`{"error":"Комната с таким именем уже есть"}`)
				}else{
					folderPath := "files/" + msg["channel"]
					os.MkdirAll(folderPath, os.ModePerm)
					h.clients[msg["channel"]] = make(map[*Client]bool)
				}
			}
			for client := range h.clients["MainChannel"] {
				select {
					case client.send <- sender.Message:
					default:
						close(client.send)
						delete(h.clients["MainChannel"], client)
				}
			}
			
		}
	}
}
