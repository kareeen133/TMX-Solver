package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"math/big"
	"strconv"
	"time"
)

type SidSignature struct {
	Type   string
	Rnd    string
	Date   string
	KeyHex string
	SigHex string
}

type ecdsaSig struct {
	R, S *big.Int
}

func GenerateSidSignature(signingNonce string) (*SidSignature, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	rnd := "tdr_" + RandomBase62(16)
	date := strconv.FormatInt(time.Now().Unix(), 10)
	sigType := "web:ecdsa"

	spki, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, err
	}

	msg := rnd + signingNonce + date + sigType
	digest := sha256.Sum256([]byte(msg))
	r, s, err := ecdsa.Sign(rand.Reader, priv, digest[:])
	if err != nil {
		return nil, err
	}
	derSig, err := asn1.Marshal(ecdsaSig{R: r, S: s})
	if err != nil {
		return nil, err
	}

	return &SidSignature{
		Type:   sigType,
		Rnd:    rnd,
		Date:   date,
		KeyHex: hex.EncodeToString(spki),
		SigHex: hex.EncodeToString(derSig),
	}, nil
}

func RandomBase62(n int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
