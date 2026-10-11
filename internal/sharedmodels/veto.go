package sharedmodels

type Veto struct {
	Vetoer UserData `json:"vetoer"`
	Hard   bool     `json:"hard"`
}
