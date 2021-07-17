package plugins

import (
	"errors"
	"sync"
	"time"
)

var (
	errAlreadyDestroyed = errors.New("Map Already Destroyed")
)

type data struct {
	value     interface{}
	createdAt time.Time
	timestamp int64
}

type Map struct {
	*sync.Mutex
	data      map[string]data
	ttl       int64
	destroyed bool
}

// New TTL Map with seconds
func NewTTLMap(ttl int64) *Map {
	h := &Map{
		data:  map[string]data{},
		ttl:   ttl,
		Mutex: &sync.Mutex{},
	}

	go h.cleanup()

	return h

}

//Destroy map and background gorutine
func (mp *Map) Destroy() error {
	if mp.destroyed {
		return errAlreadyDestroyed
	}

	mp.Lock()
	mp.destroyed = true
	mp.Unlock()

	return nil
}

func (mp *Map) Set(key string, value interface{}) error {
	one := data{
		value:     value,
		createdAt: time.Now(),
		timestamp: time.Now().Unix() + mp.ttl,
	}

	mp.Lock()
	mp.data[key] = one
	mp.Unlock()

	return nil

}

//cleanup expired data
func (mp *Map) cleanup() {
	for !mp.destroyed {

		time.Sleep(time.Duration(mp.ttl) * time.Second)

		nowTimestamp := time.Now().Unix()

		for i := range mp.data {
			mp.Lock()

			if mp.data[i].timestamp <= nowTimestamp {
				mp.Unlock()
				mp.Del(i)
			} else {
				mp.Unlock()
			}

		}
	}
}

//Get data from map and createdAt timestamp
func (mp *Map) Get(key string) (interface{}, bool, time.Time) {
	mp.Lock()
	one, ok := mp.data[key]
	mp.Unlock()

	if ok {
		if one.timestamp <= time.Now().Unix() {
			mp.Del(key)

			return false, false, time.Time{}
		} else {
			return one.value, ok, one.createdAt
		}
	} else {
		return false, false, time.Time{}
	}

}

//Remove any data from map
func (mp *Map) Del(key string) {
	mp.Lock()
	delete(mp.data, key)
	//Force GB of old map
	if len(mp.data) == 0 {
		mp.data = make(map[string]data)
	}
	mp.Unlock()

}

//Check if the key is in the map
func (mp *Map) Exist(key string) bool {
	mp.Lock()
	_, exist := mp.data[key]
	mp.Unlock()

	return exist
}
