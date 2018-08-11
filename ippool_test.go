package remote

import "testing"

func TestXici(t *testing.T) {
	store := NewRedisIPStore("118.25.7.38:6379", "ippool_xici")
	err := InitXiCiIppool(50, false, store) //, "http://106.113.242.195:9999")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStore(t *testing.T) {
	store := NewRedisIPStore("118.25.7.38:6379", "", "ippool_xici")
	size := store.Size()
	t.Log(size)
}
