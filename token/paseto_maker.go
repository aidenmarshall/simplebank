package token

import (
	"fmt"
	"time"

	"github.com/vk-rv/pvx"
)

const pasetoVersion pvx.Version = pvx.Version4

type PasetoMaker struct {
	paseto    *pvx.ProtoV4Local
	secretKey *pvx.SymKey
}

func NewPasetoMaker(keyMaterial string) (Maker, error) {
	pasetoMaker := PasetoMaker{
		paseto:    pvx.NewPV4Local(),
		secretKey: pvx.NewSymmetricKey([]byte(keyMaterial), pvx.Version4),
	}

	return &pasetoMaker, nil
}

func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}
	fmt.Println(*maker.secretKey)
	return maker.paseto.Encrypt(maker.secretKey, payload)
}

func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := Payload{}
	err := maker.paseto.Decrypt(token, maker.secretKey).ScanClaims(&payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}
