package main

type CtosCreateRoom struct {
	Nickname string `json:"nickname"`
	Version  string `json:"version"`
}

type CtosJoinRoom struct {
	Nickname string `json:"nickname"`
	RoomUid  string `json:"roomUid"`
	Version  string `json:"version"`
}

type CtosReadyToPlay struct {
	Entities []struct {
		Type_    EntityType `json:"type"`
		Position struct {
			X int `json:"x"`
			Y int `json:"y"`
		} `json:"position"`
		Direction DirectionType `json:"direction"`
	} `json:"entities"`
}

type CtosShotAt struct {
	X int `json:"x"`
	Y int `json:"y"`
}
