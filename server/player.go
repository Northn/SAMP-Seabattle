package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

type PlayerRoleType int

const (
	PRIMARY   PlayerRoleType = 1
	SECONDARY PlayerRoleType = 2
)

type Player struct {
	remoteAddr          string
	name                string
	connectedAt         time.Time
	conn                net.Conn
	entities            []*Entity
	shotPoints          []Vec2
	room                *Room
	role                PlayerRoleType
	securityErrorsCount int
	revengeRequested    bool
}

func (pl *Player) built() bool {
	return len(pl.entities) == maxPlaceableShipsCount()
}

func (pl *Player) availableEntityTypeCount(type_ EntityType) int {
	count := ENTITY_COUNT[type_]
	for _, entity := range pl.entities {
		if entity.type_ == type_ {
			count--
		}
	}
	if count < 0 {
		count = 0
	}
	return count
}

func (pl *Player) findIntersectingEntity(entity Entity) *Entity {
	for _, thisEntity := range pl.entities {
		if thisEntity.intersects(entity) {
			return thisEntity
		}
	}
	return nil
}

func (pl *Player) addEntity(entity Entity) error {
	if !canPlaceEntityType(entity.type_) {
		return fmt.Errorf("entities with specified type: %d cant be placed", entity.type_)
	}
	if pl.availableEntityTypeCount(entity.type_) <= 0 {
		return fmt.Errorf("entities count with specified type: %d has exceeded limit", entity.type_)
	}
	if intersectingEntity := pl.findIntersectingEntity(entity); intersectingEntity != nil {
		return fmt.Errorf("entity at %+v intersects with entity at %+v", intersectingEntity.position, entity.position)
	}

	pl.entities = append(pl.entities, &entity)
	return nil
}

func (pl *Player) clearEntities() {
	// self.entities = make([]Entity, maxPlaceableShipsCount())
	pl.entities = nil
	pl.shotPoints = nil
}

func (pl *Player) isAlreadyShotAt(point Vec2) bool {
	for _, thisPoint := range pl.shotPoints {
		if thisPoint.equals(point) {
			return true
		}
	}
	return false
}

func (pl *Player) shotAt(point Vec2) bool {
	if !pl.isInRoom() || !pl.room.playing() || pl.isAlreadyShotAt(point) {
		return false
	}

	pl.shotPoints = append(pl.shotPoints, point)

	sendEvent := StocAddEntity{
		Role: pl.role,
	}

	sendEvent.Entity.Position.X = point.x
	sendEvent.Entity.Position.Y = point.y
	sendEvent.Entity.Direction = HORIZONTAL

	destroyed := false
	for _, entity := range pl.entities {
		if entity.destroyAtAbs(point) {
			destroyed = true
			if entity.destroyed() {
				{ // Drawing grey empty cells
					entityDimensions := entity.dimensions()
					entityDimensions.start.sub(1)
					entityDimensions.end.add(1)
					for x := entityDimensions.start.x; x <= entityDimensions.end.x; x++ {
						for y := entityDimensions.start.y; y <= entityDimensions.end.y; y++ {
							if !(x >= 1 && x <= 10 && y >= 1 && y <= 10) {
								continue
							}

							thisPoint := Vec2{x: x, y: y}
							if !pl.isAlreadyShotAt(thisPoint) {
								sendEvent := StocAddEntity{
									Role: pl.role,
								}

								sendEvent.Entity.Type_ = EMPTY_CELL
								sendEvent.Entity.Position.X = thisPoint.x
								sendEvent.Entity.Position.Y = thisPoint.y
								sendEvent.Entity.Direction = HORIZONTAL

								pl.shotPoints = append(pl.shotPoints, thisPoint)

								pl.announceToRoom(ADD_ENTITY, sendEvent)
							}
						}
					}
				}

				{ // Remove any entities located at our entity position
					entityDimensions := entity.dimensions()
					sendEvent := StocClearBattlefield{
						Role: pl.role,
					}
					sendEvent.Start.X = entityDimensions.start.x
					sendEvent.Start.Y = entityDimensions.start.y
					sendEvent.End.X = entityDimensions.end.x
					sendEvent.End.Y = entityDimensions.end.y
					pl.enemy().send(CLEAR_BATTLEFIELD, sendEvent)
				}

				{ // Add entity to map itself for enemy
					sendEvent := StocAddEntity{
						Role: pl.role,
					}

					sendEvent.Entity.Type_ = entity.type_
					sendEvent.Entity.Position.X = entity.position.x
					sendEvent.Entity.Position.Y = entity.position.y
					sendEvent.Entity.Direction = entity.direction

					pl.enemy().send(ADD_ENTITY, sendEvent)
				}

				{ // Let this player let know that his ship was destroyed
					sendEvent.Entity.Type_ = X_MARK
					pl.send(ADD_ENTITY, sendEvent)
				}
				return false // do not switch sides
			}
			break
		}
	}

	if !destroyed {
		sendEvent.Entity.Type_ = EMPTY_CELL
	} else {
		sendEvent.Entity.Type_ = X_MARK
	}
	pl.announceToRoom(ADD_ENTITY, sendEvent)
	return !destroyed // we should not switch turn if we made a correct shot
}

func (pl *Player) isTotallyDead() bool {
	for _, entity := range pl.entities {
		if !entity.destroyed() {
			return false
		}
	}
	return true
}

func (pl *Player) enemy() *Player {
	if !pl.isInRoom() {
		return nil
	}

	enemy := pl.room.primary
	if pl.role == PRIMARY {
		enemy = pl.room.secondary
	}
	return enemy
}

func (pl *Player) logInfo(format string, args ...any) {
	str := fmt.Sprintf("[%s]: %s\n", pl.remoteAddr, format)
	log.Printf(str, args...)
}

func (pl *Player) logPanic(format string, args ...any) {
	str := fmt.Sprintf("[%s]: PANIC: %s\n", pl.remoteAddr, format)
	log.Panicf(str, args...)
}

func (pl *Player) send(code EventCode, data any) {
	if pl.conn == nil {
		return
	}

	response := Event{}
	response.Code = code
	if data != nil {
		dataIn, err := json.Marshal(data)
		if err != nil {
			pl.logPanic("%s", err)
		}
		response.Data = json.RawMessage(dataIn)
	}

	responseStr, err := json.Marshal(response)
	if err != nil {
		pl.logPanic("%s", err)
	}
	responseStr = append(responseStr, '\n')

	pl.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	pl.conn.Write(responseStr)
}

func (pl *Player) unknownError(format string, args ...any) {
	str := fmt.Sprintf(format, args...)
	pl.logInfo(str)
	pl.send(UNKNOWN_ERROR, StocUnknownError{
		Error: str,
	})
}

func (pl *Player) securityError(format string, args ...any) {
	str := fmt.Sprintf(format, args...)
	pl.logInfo(str)
	pl.send(SECURITY_ERROR, StocUnknownError{
		Error: str,
	})

	pl.securityErrorsCount++
	if MAX_SECURITY_ERRORS_COUNT != 0 && pl.securityErrorsCount >= MAX_SECURITY_ERRORS_COUNT {
		pl.disconnect()
	}
}

func (pl *Player) destroy() {
	if pl.isInRoom() {
		pl.room.announce(PLAYER_DISCONNECTED, StocPlayerDisconnected{
			Role: pl.role,
		})
		pl.room.destroy()
	}
	pl.disconnect()
}

func (pl *Player) announceToRoom(code EventCode, data any) error {
	if !pl.isInRoom() {
		return errors.New("player is not in any room")
	}
	pl.room.announce(code, data)
	return nil
}

func (pl *Player) canMakeMove() bool {
	return pl.isInRoom() && pl.room.turn == pl.role
}

func (pl *Player) isInRoom() bool {
	return pl.room != nil
}

func (pl *Player) disconnect() {
	if pl.conn == nil {
		return
	}

	pl.send(DISCONNECT, nil)
	pl.conn.Close()
	pl.conn = nil
}
