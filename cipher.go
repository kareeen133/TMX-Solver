package main

import (
	"encoding/hex"
	"fmt"
)

func XorCyclic(key, ct []byte) []byte {
	out := make([]byte, len(ct))
	for i := range ct {
		out[i] = ct[i] ^ key[i%len(key)]
	}
	return out
}

func XorCyclicStr(key, ct string) string {
	return string(XorCyclic([]byte(key), []byte(ct)))
}

func DecodeTdz(arg string) (string, error) {
	if len(arg) < 32 {
		return "", fmt.Errorf("tdz arg too short: %d", len(arg))
	}
	key := arg[:32]
	hexCt := arg[32:]
	bs, err := hex.DecodeString(hexCt)
	if err != nil {
		return "", fmt.Errorf("hex decode: %w", err)
	}
	return string(XorCyclic([]byte(key), bs)), nil
}

func EncodeTdz(key, plaintext string) string {
	if len(key) < 32 {
		k := key + key
		for len(k) < 32 {
			k += k
		}
		key = k[:32]
	}
	ct := XorCyclic([]byte(key[:32]), []byte(plaintext))
	return key[:32] + hex.EncodeToString(ct)
}

func HexEncode(b []byte) string { return hex.EncodeToString(b) }

func HexDecode(s string) ([]byte, error) { return hex.DecodeString(s) }
