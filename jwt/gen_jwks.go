package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"os"
)

// rfc7517
type jwksSpec struct {
	Keys []keySpec `json:"keys"`
}

type keySpec struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Use string `json:"use"`
	Kid string `json:"kid"`
}

func main() {
	keyBytes, err := ioutil.ReadFile("jwt/key.pub")
	if err != nil {
		panic(err)
	}
	key, err := jwt.ParseECPublicKeyFromPEM(keyBytes)
	keys := jwksSpec{
		Keys: []keySpec{
			{
				Kty: "EC",
				Crv: key.Params().Name,
				X:   base64.RawURLEncoding.EncodeToString((*key.X).Bytes()),
				Y:   base64.RawURLEncoding.EncodeToString((*key.Y).Bytes()),
				Use: "enc",
				Kid: "1",
			},
		},
	}
	data, err := json.Marshal(&keys)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("jwt/jwks.json", data, os.ModePerm); err != nil {
		panic(err)
	}
}
