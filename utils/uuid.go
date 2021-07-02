package utils

import (
	"github.com/google/uuid"
)

// IsValidUUID indicates whether the string as argument contains a valid UUID
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
