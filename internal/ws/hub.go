// hub.go
package ws

import "sync"

type Hub struct {
	muRoom sync.Mutex
	Rooms  map[string]*Room
}

func NewHub() *Hub {
	return &Hub{
		Rooms: make(map[string]*Room),
	}
}

func (h *Hub) GetOrCreate(roomId string) *Room {
	h.muRoom.Lock()
	defer h.muRoom.Unlock()

	room, ok := h.Rooms[roomId]

	if ok {
		return room
	}

	room = NewRooms(h, roomId)
	h.Rooms[roomId] = room

	go room.Run()

	return room
}

func (h *Hub) GetRoom(roomId string) *Room {

	h.muRoom.Lock()
	defer h.muRoom.Unlock()

	room, ok := h.Rooms[roomId]

	if ok {
		return room
	}

	return nil
}

func (h *Hub) removeRoom(roomId string) {

	h.muRoom.Lock()
	defer h.muRoom.Unlock()

	_, ok := h.Rooms[roomId]

	if ok {
		delete(h.Rooms, roomId)
		return

	}

}
