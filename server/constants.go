package main

import "time"

// Versioning
const (
	SERVER_VERSION          = "1.0.0"
	CLIENT_VERSION_REQUIRED = "1.0.0"
)

// TCP server configuration
const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = 5691
	CONN_TYPE = "tcp"
)

// Ships types
// Note that all these constant values must equal to client-side values!
const (
	// System types
	X_MARK     EntityType = 1 // Should be sent only to spawn X mark on opponent battlefield to hide ship position, type before destroying
	EMPTY_CELL EntityType = 2 // DNM

	FOURDECK   EntityType = 3
	THREEDECK  EntityType = 4
	DOUBLEDECK EntityType = 5
	SINGLEDECK EntityType = 6
)

// Ships dimensions (in horizontal)
// Note that all these constant values must equal to client-side values!
var ENTITY_SIZE = map[EntityType]Vec2{
	FOURDECK:   {x: 4, y: 1},
	THREEDECK:  {x: 3, y: 1},
	DOUBLEDECK: {x: 2, y: 1},
	SINGLEDECK: {x: 1, y: 1},
}

// Ships count limits. Will be used also to check if entity is placeable
// Note that all these constant values must equal to client-side values!
var ENTITY_COUNT = map[EntityType]int{
	FOURDECK:   1,
	THREEDECK:  2,
	DOUBLEDECK: 3,
	SINGLEDECK: 4,
}

const (
	MAX_INITIAL_TIMEOUT         = 10 * time.Minute
	MAX_BUILDING_TIMEOUT        = 10 * time.Minute
	MAX_GAMEPLAY_TIMEOUT        = 1 * time.Hour
	MAX_REVENGE_REQUEST_TIMEOUT = 3 * time.Minute
)

// Handshake and ping timeout
const (
	MAX_PING_TIMEOUT      = 600 * time.Second
	MAX_HANDSHAKE_TIMEOUT = 5 * time.Second
)

// Nickname validation requirements
const (
	MIN_NICKNAME_LEN          = 3
	MAX_NICKNAME_LEN          = 20
	NICKNAME_VALIDATION_REGEX = `^[a-zA-Z0-9А-Яа-я _]*$`
)

// Antiflood
// Will kick if sent MAX_EVENTS_COUNT events within MIN_EVENTS_INTERVAL
const (
	MIN_EVENTS_INTERVAL = 150 * time.Millisecond
	MAX_EVENTS_COUNT    = 5
)

const MAX_SECURITY_ERRORS_COUNT = 0 // 0 = unlimited
