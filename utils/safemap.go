package utils

import (
	"sync"
)

type SafeMap struct {
	lock *sync.RWMutex
	mp   map[interface{}]interface{}
}

// NewSafeMap return new safemap.
func NewSafeMap() *SafeMap {
	return &SafeMap{
		lock: new(sync.RWMutex),
		mp:   make(map[interface{}]interface{}),
	}
}

// Get from maps return the k's value.
func (this *SafeMap) Get(k interface{}) interface{} {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if val, ok := this.mp[k]; ok {
		return val
	}
	return nil
}

// Maps the given key and value.
// Returns false if the key is already in the map and changes nothing.
func (this *SafeMap) Set(k interface{}, v interface{}) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if val, ok := this.mp[k]; !ok {
		this.mp[k] = v
	} else if val != v {
		this.mp[k] = v
	} else {
		return false
	}
	return true
}

// Returns true if k is exist in the map.
func (this *SafeMap) Check(k interface{}) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if _, ok := this.mp[k]; !ok {
		return false
	}
	return true
}

// Delete the given key and value.
func (this *SafeMap) Delete(k interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.mp, k)
}

// Items returns all items in safemap
func (this *SafeMap) Items() map[interface{}]interface{} {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.mp
}
