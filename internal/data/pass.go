package data

type Secret struct {
	Id       string `json:"id"`
	Comment  string `json:"comment"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SecretData struct {
	Id      string   `json:"id"`
	Title   string   `json:"title"`
	Secrets []Secret `json:"secrets"`
}

type PasswordData struct {
	Passwords []SecretData `json:"passwords"`
}

type PasswordBundle struct {
	PersonId string       `json:"personId"`
	Password string       `json:"password"`
	Payload  PasswordData `json:"payload"`
}

type PasswordPostInput struct {
	PersonId string
	Password string
}
