package domain

type UserID struct {
	value string
}

func (id UserID) String() string {
    return id.value
}
