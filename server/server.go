package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	listener, err := net.Listen(CONN_TYPE, fmt.Sprintf("%s:%d", CONN_HOST, CONN_PORT))
	if err != nil {
		log.Panicln(err)
	}
	defer listener.Close()
	log.Printf("Seabattle server v%s started!\n", SERVER_VERSION)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panicln(err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	player := Player{
		conn:          conn,
		connectedAt:   time.Now(),
		remoteAddr:    conn.RemoteAddr().String(),
		lastEventTime: time.Now(),
	}

	defer func() {
		player.logInfo("Closed remote connection")
		player.destroy()
	}()

	player.logInfo("New incoming connection")

	decoder := json.NewDecoder(conn)

	for {
		if !handleRequest(&player, decoder) {
			return
		}
	}
}

func handleRequest(player *Player, decoder *json.Decoder) bool {
	lastEventTime := &player.lastEventTime
	eventsCount := &player.eventsCount
	rooms := &ROOMS_CONTAINER
	timeout := MAX_PING_TIMEOUT
	if !player.isInRoom() { // No room = no handshake
		timeout = MAX_HANDSHAKE_TIMEOUT
	}

	player.conn.SetReadDeadline(time.Now().Add(timeout))

	if time.Since(*lastEventTime) <= MIN_EVENTS_INTERVAL {
		*eventsCount++
	} else {
		*eventsCount = 0
	}
	*lastEventTime = time.Now()

	if *eventsCount >= MAX_EVENTS_COUNT {
		// Antiflood
		// TODO: send notification
		return false
	}

	event := Event{}
	{ // Decode user input and close connection if invalid data
		err := decoder.Decode(&event)

		if errors.Is(err, io.EOF) {
			//logInfo("EOF exceeded. Probably unexpected disconnect by client")
			// There is also exists wsarecv: An existing connection was forcibly closed by the remote host. but it can be handled via panic ig
			return false
		}

		if errors.Is(err, os.ErrDeadlineExceeded) {
			player.logInfo("Handshake/ping timeout exceeded: %s. Initial handshaked: %t", timeout.String(), player.room != nil)
			return false
		}

		if !player.isInRoom() && (err != nil || (event.Code != CREATE_ROOM && event.Code != JOIN_ROOM)) {
			player.logInfo("Incorrect initial handshake event")
			return false
		}

		if err != nil {
			player.logInfo("Input read error: %s", err)
			return false
		}
	}

	player.mtx.Lock()
	defer player.mtx.Unlock()
	if player.isInRoom() {
		player.room.mtx.Lock()
		defer player.room.mtx.Unlock()
	}

	if player.isInRoom() &&
		(player.room.isInitialTimeoutExceeded() ||
			player.room.isGameplayTimeoutExceeded() ||
			player.room.isBuildingTimeoutExceeded() ||
			player.room.isOverTimeoutExceeded()) {
		player.announceToRoom(GAMEPLAY_TIMEOUT_EXCEEDED, nil)
		return false // break connection to let rooms and other room participants to destroy
	}

	if event.Code == PING {
		player.send(PING, nil)
		return true
	} else if event.Code == DISCONNECT {
		return false
	}

	var data interface{}
	switch event.Code {
	case CREATE_ROOM:
		data = new(CtosCreateRoom)
	case JOIN_ROOM:
		data = new(CtosJoinRoom)
	case READY_TO_PLAY:
		data = new(CtosReadyToPlay)
	case SHOT_AT:
		data = new(CtosShotAt)
	}

	if data != nil {
		err := json.Unmarshal(event.Data, data)
		if err != nil {
			player.unknownError("Data input read error: %s", err)
			return true
		}
	}

	switch event.Code {
	case CREATE_ROOM:
		if player.isInRoom() {
			player.unknownError("you are already in room %s", player.room.uid)
			return true
		}
		data := data.(*CtosCreateRoom)
		if data.Version != CLIENT_VERSION_REQUIRED {
			player.send(INVALID_CLIENT_VERSION, nil)
			return false
		}
		if !isValidNickname(data.Nickname) {
			player.send(INVALID_NICKNAME, nil)
			return true
		}
		player.name = data.Nickname

		player.send(CREATE_ROOM, StocCreateRoom{
			RoomUid: createRoom(player).uid,
		})
	case JOIN_ROOM:
		if player.isInRoom() {
			player.unknownError("you are already in room %s", player.room.uid)
			return true
		}

		data := data.(*CtosJoinRoom)
		if data.Version != CLIENT_VERSION_REQUIRED {
			player.send(INVALID_CLIENT_VERSION, nil)
			return false
		}
		if !isValidNickname(data.Nickname) {
			player.send(INVALID_NICKNAME, nil)
			return true
		}
		player.name = data.Nickname

		if room, ok := rooms.Load(data.RoomUid); ok {
			if room := room.(*Room); !room.addSecondary(player) {
				player.send(ROOM_IS_FULL, nil)
				return true
			}
		} else {
			player.send(INVALID_ROOM_UID, nil)
			return true
		}
	case READY_TO_PLAY:
		if !player.room.building() {
			player.unknownError("not in building stage")
			return true
		}
		if player.built() {
			player.unknownError("you've already built your battlefield")
			return true
		}
		data := data.(*CtosReadyToPlay)
		var lastError error
		for _, entity := range data.Entities {
			addingEntity, err := newEntity(
				entity.Type_,
				Vec2{x: entity.Position.X, y: entity.Position.Y},
				entity.Direction,
			)

			if err != nil {
				// Incorrect entity position/direction, probably a hack attempt
				lastError = err
				break
			}

			if err := player.addEntity(addingEntity); err != nil {
				// Invalid entity type/entities count limit exceeded/intersecting entities were found, probably a hack attempt
				lastError = err
				break
			}
		}

		if lastError != nil {
			player.securityError(lastError.Error())
			player.clearEntities()
		} else if player.built() {
			player.announceToRoom(READY_TO_PLAY, StocPlayerReadyToPlay{
				Role: player.role,
			})

			// Let's sync player battlefield with the server data

			{
				sendEvent := StocClearBattlefield{
					Role: player.role,
				}
				sendEvent.Start.X = 1
				sendEvent.Start.Y = 1
				sendEvent.End.X = 10
				sendEvent.End.Y = 10

				player.send(CLEAR_BATTLEFIELD, sendEvent)
			}

			for _, entity := range player.entities {
				event := StocAddEntity{
					Role: player.role,
				}
				event.Entity.Type_ = entity.type_
				event.Entity.Position.X = entity.position.x
				event.Entity.Position.Y = entity.position.y
				event.Entity.Direction = entity.direction
				player.send(ADD_ENTITY, event)
			}

			player.room.startPlaying()
		} else {
			player.unknownError("not enough entities were placed")
			player.clearEntities()
		}
	case SHOT_AT:
		if !player.room.playing() {
			player.unknownError("not in playing stage")
			return true
		}
		if !player.canMakeMove() {
			player.unknownError("not your turn")
			return true
		}

		data := data.(*CtosShotAt)

		if player.enemy().shotAt(Vec2{x: data.X, y: data.Y}) {
			player.room.switchTurn()
		}

		if player.enemy().isTotallyDead() {
			player.room.announce(PLAYER_WIN, StocPlayerWin{
				Role: player.role,
			})
			player.room.setGamestate(OVER)
		}
	case REVENGE_REQUESTED:
		if !player.room.over() {
			player.unknownError("not in over stage")
			return true
		}

		if player.revengeRequested {
			player.unknownError("you've already requested a revenge")
			return true
		}
		player.revengeRequested = true
		player.room.announce(REVENGE_REQUESTED, StocRevengeRequested{
			Role: player.role,
		})

		player.room.startRevenge()
	}
	return true
}
