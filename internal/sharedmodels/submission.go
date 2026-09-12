package sharedmodels

import (
	"time"
)

type Submission struct {
	DateCrated time.Time `json:"date_created"`
	Author     UserData  `json:"author"`
	Content    string    `json:"content"`
	Hash       string    `json:"hash"`
}
