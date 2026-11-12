package publishingservice

import "revit/internal/sharedmodels"

type PublishingService interface {
	Publish(sharedmodels.Submission) error
}
