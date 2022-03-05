package app

import (
	"sync"
)

type user struct {
	Name string
	Id   uint32
}

type userPool struct {
	m        sync.Mutex
	idToUser []*user
}

func (userPool *userPool) getUser(id uint32) (*user, bool) {
	userPool.m.Lock()
	defer userPool.m.Unlock()
	if id < 0 || id >= uint32(len(userPool.idToUser)) {
		return nil, false
	}
	return userPool.idToUser[id], true
}

func (userPool *userPool) addUser(user *user) *user {
	userPool.m.Lock()
	defer userPool.m.Unlock()
	user.Id = uint32(len(userPool.idToUser))
	userPool.idToUser = append(userPool.idToUser, user)
	return user
}
