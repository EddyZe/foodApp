package redisutil

import (
	"fmt"
)

func GenerateKey(startKey, value string) string {
	return fmt.Sprintf("%s:%s", startKey, value)
}
