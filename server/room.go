package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Gamestate int

const (
	INITIAL  Gamestate = 1
	BUILDING Gamestate = 2
	PLAYING  Gamestate = 3
	OVER     Gamestate = 4
)

type Room struct {
	mtx              sync.Mutex
	lastGamestateSet time.Time
	gamestate        Gamestate
	uid              string
	primary          *Player
	secondary        *Player
	turn             PlayerRoleType
}

func createRoom(player *Player) *Room {
	room := Room{
		lastGamestateSet: time.Now(),
		uid:              uuid.New().String(),
		gamestate:        INITIAL,
		turn:             SECONDARY, // Initial has to be SECONDARY so switchTurn() will start from PRIMARY
	}

	player.room = &room
	player.role = PRIMARY
	room.primary = player

	ROOMS_CONTAINER.Store(room.uid, &room)

	room.logInfo("Created! Requested by: %s", player.name)
	return &room
}

func (room *Room) addSecondary(player *Player) bool {
	if room.secondary == nil {
		player.room = room
		player.role = SECONDARY
		room.secondary = player

		room.announce(JOIN_ROOM, StocJoinRoom{
			PrimaryName:   room.primary.name,
			SecondaryName: room.secondary.name,
		})

		room.logInfo("Joined as secondary player: %s", player.name)

		room.startBuilding()

		return true
	}
	return false
}

func (room *Room) announce(code EventCode, data any) {
	if !room.valid() {
		return
	}
	if room.primary != nil {
		room.primary.send(code, data)
	}
	if room.secondary != nil {
		room.secondary.send(code, data)
	}
}

func (room *Room) logInfo(format string, args ...any) {
	str := fmt.Sprintf("[%s]: %s\n", room.uid, format)
	log.Printf(str, args...)
}

func (room *Room) isInitialTimeoutExceeded() bool {
	return !room.building() && !room.playing() && time.Since(room.lastGamestateSet) >= MAX_INITIAL_TIMEOUT
}

func (room *Room) isBuildingTimeoutExceeded() bool {
	return room.building() && time.Since(room.lastGamestateSet) >= MAX_BUILDING_TIMEOUT
}

func (room *Room) isGameplayTimeoutExceeded() bool {
	return room.playing() && time.Since(room.lastGamestateSet) >= MAX_GAMEPLAY_TIMEOUT
}

func (room *Room) isOverTimeoutExceeded() bool {
	return room.playing() && time.Since(room.lastGamestateSet) >= MAX_REVENGE_REQUEST_TIMEOUT
}

func (room *Room) building() bool {
	return room.gamestate == BUILDING
}

func (room *Room) playing() bool {
	return room.gamestate == PLAYING
}

func (room *Room) over() bool {
	return room.gamestate == OVER
}

func (room *Room) startRevenge() bool {
	if !room.over() || !room.primary.revengeRequested || !room.secondary.revengeRequested {
		return false
	}

	room.gamestate = INITIAL
	room.primary.revengeRequested = false
	room.secondary.revengeRequested = false
	room.startBuilding()

	room.logInfo("Revenge!")

	return true
}

func (room *Room) startBuilding() bool {
	if room.gamestate != INITIAL {
		return false
	}

	room.setGamestate(BUILDING)
	room.primary.clearEntities()
	room.secondary.clearEntities()

	room.logInfo("Building stage has started")

	return true
}

func (room *Room) setGamestate(state Gamestate) {
	room.gamestate = state
	room.lastGamestateSet = time.Now()
	room.announce(SET_GAMESTATE, StocSetGamestate{
		Gamestate_: state,
	})
}

func (room *Room) canStartPlaying() bool {
	if !room.building() || room.primary == nil || room.secondary == nil {
		return false
	}

	if !room.primary.built() || !room.secondary.built() {
		return false
	}

	return true
}

func (room *Room) startPlaying() bool {
	if !room.canStartPlaying() {
		return false
	}

	room.setGamestate(PLAYING)

	room.logInfo("Let the greatest battle begin!")

	room.switchTurn()

	return true
}

func (room *Room) switchTurn() {
	if room.turn == PRIMARY {
		room.turn = SECONDARY
	} else {
		room.turn = PRIMARY
	}

	room.announce(SET_TURN, StocSetTurn{
		Role: room.turn,
	})
}

func (room *Room) destroy() {
	room.mtx.Lock()
	defer room.mtx.Unlock()

	if !room.valid() {
		return
	}

	room.announce(ROOM_CLOSED, nil)

	removeFromRoom(&room.primary)
	removeFromRoom(&room.secondary)

	ROOMS_CONTAINER.Delete(room.uid)
	room.logInfo("Destroyed!")
}

func (room *Room) valid() bool {
	_, ok := ROOMS_CONTAINER.Load(room.uid)
	return ok
}

func removeFromRoom(field **Player) {
	if (*field) != nil {
		(*field).disconnect()
		(*field).room = nil
		(*field) = nil
	}
}

var ROOMS_CONTAINER = sync.Map{}
