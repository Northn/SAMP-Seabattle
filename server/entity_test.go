package main

import (
	"testing"
)

func TestSummonBoundaries(t *testing.T) {
	_, err1 := newEntity(FOURDECK, Vec2{x: 1, y: 3}, VERTICAL)
	_, err2 := newEntity(FOURDECK, Vec2{x: 8, y: 1}, HORIZONTAL)
	_, err3 := newEntity(FOURDECK, Vec2{x: 6, y: 4}, VERTICAL)
	_, err4 := newEntity(FOURDECK, Vec2{x: 6, y: 4}, DirectionType(3))

	if err1 == nil || err2 == nil {
		t.Errorf("Entities are out of boundaries but were summoned: %s %s", err1, err2)
	}

	if err3 != nil {
		t.Errorf("Entity is not out of boundaries but was not summoned: %s", err3)
	}

	if err4 == nil {
		t.Errorf("Entity has invalid direction but was summoned: %s", err3)
	}
}

func TestDimensions(t *testing.T) {
	startPos := Vec2{x: 5, y: 5}

	entity, _ := newEntity(FOURDECK, startPos, VERTICAL)
	dimensions := entity.dimensions()

	if !(dimensions.start.x == 5 && dimensions.start.y == 2 && dimensions.end.x == 5 && dimensions.end.y == 5) {
		t.Errorf("Entity dimensions have unexpected values")
	}
}

func TestIntersection(t *testing.T) {
	{
		entity1, _ := newEntity(DOUBLEDECK, Vec2{x: 2, y: 2}, HORIZONTAL)
		entity2, _ := newEntity(FOURDECK, Vec2{x: 5, y: 5}, VERTICAL)

		if entity1.intersects(entity2) {
			t.Errorf("Entities intersect but should not")
		}
	}

	{
		entity1, _ := newEntity(DOUBLEDECK, Vec2{x: 2, y: 5}, HORIZONTAL)
		entity2, _ := newEntity(FOURDECK, Vec2{x: 4, y: 4}, VERTICAL)

		if !entity1.intersects(entity2) {
			t.Errorf("Entities do not intersect but have to")
		}
	}

	{
		entity1, _ := newEntity(DOUBLEDECK, Vec2{x: 2, y: 5}, HORIZONTAL)
		entity2, _ := newEntity(FOURDECK, Vec2{x: 6, y: 4}, VERTICAL)

		if entity1.intersects(entity2) {
			t.Errorf("Entities intersect but should not")
		}
	}

	{
		entity1, _ := newEntity(THREEDECK, Vec2{x: 5, y: 8}, HORIZONTAL)
		entity2, _ := newEntity(THREEDECK, Vec2{x: 2, y: 10}, HORIZONTAL)

		if entity1.intersects(entity2) {
			t.Errorf("Entities intersect but should not")
		}
	}
}

func TestHorizontalAbsToLocalPoint(t *testing.T) {
	entity, _ := newEntity(THREEDECK, Vec2{x: 3, y: 4}, HORIZONTAL)

	for i := 0; i <= 2; i++ {
		point, err := entity.convertAbsPointToLocal(Vec2{x: 3 + i, y: 4})
		if err != nil {
			t.Errorf("Could not convert absolute point to local: %s", err)
		}

		if !(point.x == 1+i && point.y == 1) {
			t.Errorf("Incorrect conversion of absolute to local point")
		}
	}

	{
		point, err := entity.convertAbsPointToLocal(Vec2{x: 2, y: 5})
		if err == nil {
			t.Errorf("Conversion was done successfully even if point is out of bounds: %+v", point)
		}
	}
}

func TestVerticalAbsToLocalPoint(t *testing.T) {
	entity, _ := newEntity(THREEDECK, Vec2{x: 3, y: 4}, VERTICAL)

	for i := 0; i <= 2; i++ {
		point, err := entity.convertAbsPointToLocal(Vec2{x: 3, y: 2 + i})
		if err != nil {
			t.Errorf("Could not convert absolute point to local: %s", err)
		}

		if !(point.x == 1 && point.y == 1+i) {
			t.Errorf("Incorrect conversion of absolute to local point")
		}
	}

	{
		point, err := entity.convertAbsPointToLocal(Vec2{x: 2, y: 5})
		if err == nil {
			t.Errorf("Conversion was done successfully even if point is out of bounds: %+v", point)
		}
	}
}

func TestDestroy(t *testing.T) {
	entity, _ := newEntity(THREEDECK, Vec2{x: 3, y: 4}, VERTICAL)

	if entity.destroyAtAbs(Vec2{x: 5, y: 5}) {
		t.Errorf("Entity point was destroyed but should not")
	}

	if !entity.destroyAtAbs(Vec2{x: 3, y: 2}) {
		t.Errorf("Entity point was not destroyed but has to be")
	}

	// Destroy again
	if entity.destroyAtAbs(Vec2{x: 3, y: 2}) {
		t.Errorf("Entity point already was destroyed but it's destroyed again")
	}

	if entity.destroyed() {
		t.Errorf("Entity is destroyed but can't be")
	}
	entity.destroyAtAbs(Vec2{x: 3, y: 2})
	entity.destroyAtAbs(Vec2{x: 3, y: 3})
	entity.destroyAtAbs(Vec2{x: 3, y: 4})
	if !entity.destroyed() {
		t.Errorf("Entity is not destroyed but has to be")
	}
}
