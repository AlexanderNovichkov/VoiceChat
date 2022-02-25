package app

import (
	"sync"
)

type User struct {
	Name string
	Id   uint32
}

type UserPool struct {
	m        sync.Mutex
	idToUser []*User
}

func (userPool *UserPool) GetUser(id uint32) (*User, bool) {
	userPool.m.Lock()
	defer userPool.m.Unlock()
	if id < 0 || id >= uint32(len(userPool.idToUser)) {
		return nil, false
	}
	return userPool.idToUser[id], true
}

func (userPool *UserPool) AddUser(user *User) *User {
	userPool.m.Lock()
	defer userPool.m.Unlock()
	user.Id = uint32(len(userPool.idToUser))
	userPool.idToUser = append(userPool.idToUser, user)
	return user
}
