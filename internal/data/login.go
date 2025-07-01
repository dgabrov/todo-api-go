package data

type LoginData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type TokenPersonData struct {
	Login      string `json:"login"`
	Name       string `json:"name"`
	ProvidedId string `json:"providedId"`
	PersonId   string `json:"personId"`
	Token      string `json:"token"`
}

type LoginAuthData struct {
	Id     string   `json:"id"`
	Name   string   `json:"name"`
	Login  string   `json:"login"`
	Rights []string `json:"rights"`
}

type PersonIdData struct {
	PersonId string `json:"personId"`
}
