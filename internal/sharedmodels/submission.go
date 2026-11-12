package sharedmodels

import (
	"revit/internal/db"
	"time"
)

type Submission struct {
	DateCrated time.Time `json:"date_created"`
	Author     UserData  `json:"author"`
	Content    string    `json:"content"`
	Hash       string    `json:"hash"`
}

func MapDBSubmissionToModel(submission db.Submission, author UserData) Submission {
	return Submission{DateCrated: submission.CreatedAt.Time, Author: author, Content: submission.Content, Hash: submission.Hash}
}
