package sharedmodels

import "revit/internal/db"

type UserData struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func MapUserDataFromDBModel(model db.User) UserData {
	var email string
	if model.Email.Valid {
		email = model.Email.String
	}

	return UserData{
		Email: email,
		Name:  model.Name,
	}
}
