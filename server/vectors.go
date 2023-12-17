package main

type Vec2 struct {
	x int
	y int
}

func (vec *Vec2) add(num int) {
	vec.x += num
	vec.y += num
}

func (vec *Vec2) sub(num int) {
	vec.x -= num
	vec.y -= num
}

func (vec Vec2) equals(another Vec2) bool {
	return vec.x == another.x && vec.y == another.y
}

type Vec2x2 struct {
	start Vec2
	end   Vec2
}

func (vec Vec2x2) intersects(another Vec2x2) bool {
	return vec.start.x <= another.end.x && vec.end.x >= another.start.x && vec.start.y <= another.end.y && vec.end.y >= another.start.y
}
