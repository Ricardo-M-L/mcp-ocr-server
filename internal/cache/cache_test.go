package cache

import (
	"testing"
	"time"
)

func TestCache_GetSet(t *testing.T) {
	cache := NewCache(10, time.Second*5, true)

	// 测试设置和获取
	cache.Set("key1", "value1")

	value, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}

	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(10, time.Millisecond*100, true)

	cache.Set("key1", "value1")

	// 等待过期
	time.Sleep(time.Millisecond * 150)

	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be expired")
	}
}

func TestCache_MaxSize(t *testing.T) {
	cache := NewCache(3, time.Hour, true)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	cache.Set("key4", "value4") // 应该淘汰 key1

	if cache.Size() > 3 {
		t.Errorf("Cache size exceeded max: %d", cache.Size())
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(10, time.Hour, true)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0, got %d", cache.Size())
	}
}

func TestGenerateKey(t *testing.T) {
	data := []byte("test data")
	key1 := GenerateKey(data)
	key2 := GenerateKey(data)

	if key1 != key2 {
		t.Error("Expected same key for same data")
	}

	key3 := GenerateKey(data, "option1")
	if key1 == key3 {
		t.Error("Expected different key with different options")
	}
}

func TestCache_Disabled(t *testing.T) {
	cache := NewCache(10, time.Hour, false)

	cache.Set("key1", "value1")

	_, found := cache.Get("key1")
	if found {
		t.Error("Expected cache to be disabled")
	}
}