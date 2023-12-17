package main

type StocUnknownError struct {
	Error string `json:"error"`
}

type StocSecurityError struct {
	Error string `json:"error"`
}

type StocCreateRoom struct {
	RoomUid string `json:"roomUid"`
}

type StocJoinRoom struct {
	PrimaryName   string `json:"primaryName"`
	SecondaryName string `json:"secondaryName"`
}

type StocPlayerDisconnected struct {
	Role PlayerRoleType `json:"role"`
}

type StocSetGamestate struct {
	Gamestate_ Gamestate `json:"gamestate"`
}

type StocPlayerReadyToPlay struct {
	Role PlayerRoleType `json:"role"`
}

type StocClearBattlefield struct {
	Role  PlayerRoleType `json:"role"`
	Start struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"start"`
	End struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"end_"`
}

type StocAddEntity struct {
	Role   PlayerRoleType `json:"role"`
	Entity struct {
		Type_    EntityType `json:"type"`
		Position struct {
			X int `json:"x"`
			Y int `json:"y"`
		} `json:"position"`
		Direction DirectionType `json:"direction"`
	} `json:"entity"`
}

type StocSetTurn struct {
	Role PlayerRoleType `json:"role"`
}

type StocPlayerWin struct {
	Role PlayerRoleType `json:"role"`
}

type StocRevengeRequested struct {
	Role PlayerRoleType `json:"role"`
}
