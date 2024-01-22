package state

import (
	"sync"
)

/**
  设置运行状态
*/

var (
	State = new(sync.Map)
)

func SetState(site string) {
	State.Store(site, true)
}

func DeleteState(site string) {
	State.Delete(site)
}

func ExistsState(state string) bool {
	_, ok := State.Load(state)
	return ok
}

func GetStateList() (l []string) {
	State.Range(func(key, _ any) bool {
		l = append(l, key.(string))
		return true
	})
	return
}
