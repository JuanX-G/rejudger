package submission

import (
	"revit/internal/config/sharedmodels"
	"revit/internal/db"
	services "revit/internal/testingservices"
)

var DEFAULT_USERDATA = sharedmodels.UserData{Email: services.DEFAULT_EMAIL, Name: services.DEFAULT_USERNAME}
var DEFAULT_RESP_SUBMISSION_1 = sharedmodels.Submission{Content: "a", Author: DEFAULT_USERDATA, Hash: "a"}
var DEFAULT_RESP_SUBMISSION_2 = sharedmodels.Submission{Content: "b", Author: DEFAULT_USERDATA, Hash: "b"}
var DEFAULT_RESP_SUBMISSIONS = []sharedmodels.Submission{DEFAULT_RESP_SUBMISSION_1, DEFAULT_RESP_SUBMISSION_2}

var DEFAULT_SUBMISSION_1 = db.Submission{ID: 1, Content: "a", Author: 1, Hash: "a"}
var DEFAULT_SUBMISSION_2 = db.Submission{ID: 2, Content: "b", Author: 1, Hash: "b"}
var DEFAULT_SUBMISSIONS = []db.Submission{DEFAULT_SUBMISSION_1, DEFAULT_SUBMISSION_2}

const TEST_SUBMISSION = "Lorem Ipsum"
