package remote

import "testing"

func TestStore(t *testing.T) {
	store := NewRedisIPStore("118.25.7.38:6379", "", "ippool_xici")
	size := store.Size()
	t.Log(size)
}
