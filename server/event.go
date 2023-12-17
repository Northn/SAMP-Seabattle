package main

import "encoding/json"

type EventCode int

type Event struct {
	Code EventCode       `json:"code"`
	Data json.RawMessage `json:"data"`
}

// System events to communicate between client and server
const ( // CTOS: Client to Server, STOC: Server to Client
	INVALID_EVENT             EventCode = 0  // STOC: data: nil
	UNKNOWN_ERROR             EventCode = 1  // STOC: see StocUnknownError
	SECURITY_ERROR            EventCode = 2  // STOC: see StocSecurityError
	PING                      EventCode = 3  // STOC: data: nil; CTOS: data: nil // Sent to let client and server know that they're both alive
	DISCONNECT                EventCode = 4  // STOC: data: nil; CTOS: data: nil // Sent before client or server disconnect to notify remote
	CREATE_ROOM               EventCode = 5  // CTOS: see CtosCreateRoom; STOC: see StocCreateRoom
	JOIN_ROOM                 EventCode = 6  // CTOS: see CtosJoinRoom; STOC: see StocJoinRoom
	ROOM_IS_FULL              EventCode = 7  // STOC: data: nil // Sent as response to JOIN_ROOM CTOS if room is full
	INVALID_NICKNAME          EventCode = 8  // STOC: data: nil // Sent if nickname is shorter than MIN_NICKNAME_LEN or longer than MAX_NICKNAME_LEN or does not qualify to NICKNAME_VALIDATION_REGEX
	INVALID_ROOM_UID          EventCode = 9  // STOC: data: nil // Sent if room does not exist
	INVALID_CLIENT_VERSION    EventCode = 10 // STOC: data: nil // Sent if invalid client version
	ROOM_CLOSED               EventCode = 11 // STOC: data: nil // Sent if room was closed by timeout or 2nd player was disconnected
	GAMEPLAY_TIMEOUT_EXCEEDED EventCode = 12 // STOC: data: nil // Sent if isInitialTimeoutExceeded() || isGameplayTimeoutExceeded() || isBuildingTimeoutExceeded()
	PLAYER_DISCONNECTED       EventCode = 13 // STOC: see StocPlayerDisconnected
	SET_GAMESTATE             EventCode = 14 // STOC: see StocSetGamestate
	READY_TO_PLAY             EventCode = 15 // CTOS: see CtosReadyToPlay; STOC: see StosReadyToPlay
	CLEAR_BATTLEFIELD         EventCode = 16 // STOC: see StocClearBattlefield // Clears all entities in area
	ADD_ENTITY                EventCode = 17 // STOC: see StocAddEntity // Adds entity to battlefield
	SET_TURN                  EventCode = 18 // STOC: see StocAddEntity // Sets current player turn
	SHOT_AT                   EventCode = 19 // CTOS: see CtosShotAt // Gameplay process itself
	PLAYER_WIN                EventCode = 20 // STOC: see StocPlayerWin // Sent when one of players are done and won
	REVENGE_REQUESTED         EventCode = 21 // STOC: see StocRevengeRequested; CTOS: data: nil
)
