package events

import "time"

type UserSwitched struct {
	name string
	time time.Time

	Username string
	Token    string
}

func (e UserSwitched) New(args ...interface{}) Event {
	return &UserSwitched{
		name:     "USER_SWITCHED",
		time:     time.Now(),
		Username: args[0].(string),
		Token:    args[1].(string),
	}
}

func (e *UserSwitched) Name() string {
	return e.name
}

func (e *UserSwitched) Time() time.Time {
	return e.time
}
