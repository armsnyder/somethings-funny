package main

import (
	"flag"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"time"
)

func main() {
	days := flag.Int64("days", 30, "token validity in days")
	aud := flag.String("aud", "", "jwt audience")
	flag.Parse()
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Unix() + *days*24*60*60,
		Issuer:    "default",
		Audience:  *aud,
	})
	token.Header["kid"] = "1"
	keyBytes, err := ioutil.ReadFile("jwt/key.pem")
	if err != nil {
		panic(err)
	}
	key, err := jwt.ParseECPrivateKeyFromPEM(keyBytes)
	if err != nil {
		panic(err)
	}
	signed, err := token.SignedString(key)
	if err != nil {
		panic(err)
	}
	fmt.Println(signed)
}
