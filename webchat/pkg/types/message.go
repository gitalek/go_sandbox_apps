package types

import "time"

// message represents a single chat message
type message struct {
	Name string
	Message string
	When time.Time
}
