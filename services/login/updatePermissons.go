package login

import (
	//"fmt"
)

type FetchPermissionsQuery struct {
	Token string `json:"token"`
	Context string `json:"context"`

}

type AddPermisionsResponse struct {
}
