package dbpublishingservice

import (
	"context"
	"revit/internal/db"
)

type dbPublishingServiceStore interface {
	GetPublishedByHash(ctx context.Context, hash string) (db.PublishedSubmission, error)
	InsertIntoPublishedSubmissions(ctx context.Context, arg db.InsertIntoPublishedSubmissionsParams) error
	GetUserById(ctx context.Context, id int64) (db.User, error)
}
