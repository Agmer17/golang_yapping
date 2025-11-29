// room.go
package ws

type Room struct {
	Id        string
	Clients   map[*Client]bool
	Broadcast chan []byte

	Register   chan *Client
	Unregister chan *Client

	Hub *Hub
}

func NewRooms(hub *Hub, id string) *Room {
	return &Room{
		Id:         id,
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Hub:        hub,
	}
}

func (r *Room) Run() {

	defer func() {
		close(r.Broadcast)
		close(r.Register)
		close(r.Unregister)
		r.Hub.removeRoom(r.Id)

	}()
	for {
		select {
		case nc := <-r.Register:
			r.Clients[nc] = true

		case uc := <-r.Unregister:
			if _, ok := r.Clients[uc]; ok {
				delete(r.Clients, uc)
				close(uc.Send)
			}

			if len(r.Clients) == 0 {
				return
			}
		case pl := <-r.Broadcast:
			for client := range r.Clients {
				select {
				case client.Send <- pl:
				default:
					close(client.Send)
					delete(r.Clients, client)
				}
			}
		}
	}
}
