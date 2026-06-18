package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"time"
)

type APIKey struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Credits   int64  `json:"credits"`
	Used      int64  `json:"used"`
	Disabled  bool   `json:"disabled"`
	CreatedAt int64  `json:"created_at"`
	LastUsed  int64  `json:"last_used"`
}

type KeyStore struct {
	mu   sync.Mutex
	path string
	keys map[string]*APIKey
}

var errNoKey = errors.New("invalid api key")
var errNoCredits = errors.New("no credits remaining")
var errKeyDisabled = errors.New("key disabled")

func NewKeyStore(path string) (*KeyStore, error) {
	ks := &KeyStore{path: path, keys: map[string]*APIKey{}}
	if path == "" {
		return ks, nil
	}
	if data, err := os.ReadFile(path); err == nil {
		var arr []*APIKey
		if json.Unmarshal(data, &arr) == nil {
			for _, k := range arr {
				ks.keys[k.Key] = k
			}
		}
	}
	return ks, nil
}

func (ks *KeyStore) flush() {
	if ks.path == "" {
		return
	}
	arr := make([]*APIKey, 0, len(ks.keys))
	for _, k := range ks.keys {
		arr = append(arr, k)
	}
	if b, err := json.MarshalIndent(arr, "", "  "); err == nil {
		os.WriteFile(ks.path, b, 0600)
	}
}

func genKey() string {
	b := make([]byte, 18)
	rand.Read(b)
	return "tmx_" + hex.EncodeToString(b)
}

func (ks *KeyStore) Create(label string, credits int64) *APIKey {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	k := &APIKey{
		Key:       genKey(),
		Label:     strings.TrimSpace(label),
		Credits:   credits,
		CreatedAt: time.Now().UnixMilli(),
	}
	ks.keys[k.Key] = k
	ks.flush()
	return k
}

func (ks *KeyStore) Adjust(key string, delta int64) (*APIKey, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	k, ok := ks.keys[key]
	if !ok {
		return nil, errNoKey
	}
	k.Credits += delta
	if k.Credits < 0 {
		k.Credits = 0
	}
	ks.flush()
	return k, nil
}

func (ks *KeyStore) SetDisabled(key string, disabled bool) (*APIKey, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	k, ok := ks.keys[key]
	if !ok {
		return nil, errNoKey
	}
	k.Disabled = disabled
	ks.flush()
	return k, nil
}

func (ks *KeyStore) Delete(key string) bool {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	if _, ok := ks.keys[key]; !ok {
		return false
	}
	delete(ks.keys, key)
	ks.flush()
	return true
}

func (ks *KeyStore) Get(key string) (*APIKey, bool) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	k, ok := ks.keys[key]
	if !ok {
		return nil, false
	}
	cp := *k
	return &cp, true
}

func (ks *KeyStore) List() []*APIKey {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	arr := make([]*APIKey, 0, len(ks.keys))
	for _, k := range ks.keys {
		cp := *k
		arr = append(arr, &cp)
	}
	return arr
}

func (ks *KeyStore) Consume(key string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	k, ok := ks.keys[key]
	if !ok {
		return errNoKey
	}
	if k.Disabled {
		return errKeyDisabled
	}
	if k.Credits <= 0 {
		return errNoCredits
	}
	k.Credits--
	k.Used++
	k.LastUsed = time.Now().UnixMilli()
	ks.flush()
	return nil
}

func (ks *KeyStore) Refund(key string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	if k, ok := ks.keys[key]; ok {
		k.Credits++
		if k.Used > 0 {
			k.Used--
		}
		ks.flush()
	}
}
