package events

import "time"

type UserRegistered struct {
	name string
	time time.Time

	Username string
	Token    string
}

func (e UserRegistered) New(args ...interface{}) Event {
	return &UserRegistered{
		name:     "USER_REGISTERED",
		time:     time.Now(),
		Username: args[0].(string),
		Token:    args[1].(string),
	}
}

func (e *UserRegistered) Name() string {
	return e.name
}

func (e *UserRegistered) Time() time.Time {
	return e.time
}
