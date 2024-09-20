package types

type User struct {
	ID    int64
	UUID  string
	Name  string
	Email string
}

type Role struct {
	ID   int64
	Name string
}
