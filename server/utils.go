package main

import "regexp"

func canPlaceEntityType(type_ EntityType) bool {
	val, ok := ENTITY_COUNT[type_]
	return ok && val > 0
}

func maxPlaceableShipsCount() int {
	count := 0
	for _, amount := range ENTITY_COUNT {
		count += amount
	}
	return count
}

func isValidNickname(nickname string) bool {
	len_ := len(nickname)
	return len_ >= MIN_NICKNAME_LEN && len_ <= MAX_NICKNAME_LEN && regexp.MustCompile(NICKNAME_VALIDATION_REGEX).MatchString(nickname)
}
