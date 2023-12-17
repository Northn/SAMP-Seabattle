package main

import (
	"errors"
	"fmt"
)

type EntityType int
type DirectionType int

const (
	HORIZONTAL DirectionType = 1
	VERTICAL   DirectionType = 2
)

type Entity struct {
	type_           EntityType
	position        Vec2
	direction       DirectionType
	destroyedPoints []Vec2
}

func newEntity(type_ EntityType, position Vec2, direction DirectionType) (Entity, error) {
	entity := Entity{
		type_:     type_,
		position:  position,
		direction: direction,
	}

	if direction != HORIZONTAL && direction != VERTICAL {
		return entity, fmt.Errorf("incorrect entity orientation: %d", direction)
	}

	dimensions := entity.dimensions()
	if !(dimensions.start.x >= 1 && dimensions.start.y >= 1 && dimensions.end.x <= 10 && dimensions.end.y <= 10) {
		return entity, fmt.Errorf("incorrect entity boundaries: type: %d, dimensions: %+v, orientation: %d", type_, dimensions, direction)
	}

	return entity, nil
}

func (ent *Entity) size() Vec2 {
	size_ := ENTITY_SIZE[ent.type_]
	x := size_.x
	y := size_.y

	if ent.direction != HORIZONTAL {
		x, y = y, x
	}
	return Vec2{x: x, y: y}
}

func (ent *Entity) dimensions() Vec2x2 {
	size_ := ent.size()

	return Vec2x2{
		start: Vec2{x: ent.position.x, y: ent.position.y - size_.y + 1},
		end:   Vec2{x: ent.position.x + size_.x - 1, y: ent.position.y},
	}
}

func (ent *Entity) intersects(anotherEntity Entity) bool {
	dimensions := ent.dimensions()
	dimensions.start.sub(1)
	dimensions.end.add(1)
	return dimensions.intersects(anotherEntity.dimensions())
}

func (ent *Entity) convertAbsPointToLocal(point Vec2) (Vec2, error) {
	size_ := ent.size()
	dimensions := ent.dimensions()

	for x := 0; x < size_.x; x++ {
		for y := 0; y < size_.y; y++ {
			if dimensions.start.x+x == point.x && dimensions.start.y+y == point.y {
				return Vec2{x: x + 1, y: y + 1}, nil
			}
		}
	}

	return Vec2{x: -1, y: -1}, errors.New("point is out of entity boundaries")
}

func (ent *Entity) destroyAt(point Vec2) bool {
	for _, thisPoint := range ent.destroyedPoints {
		if thisPoint.equals(point) {
			return false
		}
	}

	ent.destroyedPoints = append(ent.destroyedPoints, point)

	return true
}

func (ent *Entity) destroyAtAbs(point Vec2) bool {
	localPoint, err := ent.convertAbsPointToLocal(point)

	if err != nil {
		return false
	}

	return ent.destroyAt(localPoint)
}

func (ent *Entity) destroyed() bool {
	size_ := ent.size()
	return len(ent.destroyedPoints) == size_.x*size_.y
}
