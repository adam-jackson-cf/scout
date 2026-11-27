package cache

import "canary/store"

func Warm(s store.Store, key string) bool {
	_, ok := s.Get(key)
	return ok
}
