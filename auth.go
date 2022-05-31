package gobore

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"net"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type authenticator struct {
	hash.Hash
}

func newAuthenticator(secret string) *authenticator {
	return &authenticator{Hash: hmac.New(sha256.New, []byte(secret))}
}

func (a *authenticator) serverHandshake(conn net.Conn) error {
	challenge := uuid.New().String()
	err := sendJson(conn, serverMessage{Type: serverMessageChallenge, Challenge: challenge})
	if err != nil {
		zap.L().Error("Failed to sendJson", zap.Error(err))
		return err
	}

	cm := clientMessage{}
	err = recvJsonWithTimeout(conn, &cm)
	if err != nil {
		zap.L().Error("Failed to recvJsonWithTimeout", zap.Error(err))
		return err
	}
	if !a.validate(challenge, cm.Authenticate) {
		return errors.New("invalid secret")
	}
	return nil
}

func (a *authenticator) clientHandshake(conn net.Conn) error {
	sm := serverMessage{}
	err := recvJsonWithTimeout(conn, &sm)
	if err != nil {
		zap.L().Error("Failed to recvJsonWithTimeout", zap.Error(err))
		return err
	}

	cm := clientMessage{Type: clientMessageAuthenticate}
	cm.Authenticate = a.answer(sm.Challenge)
	return sendJson(conn, cm)
}

func (a *authenticator) validate(challenge string, tag string) bool {
	a.Reset()
	_, _ = a.Write([]byte(challenge))
	return hex.EncodeToString(a.Sum(nil)) == tag
}

func (a *authenticator) answer(challenge string) string {
	a.Reset()
	_, _ = a.Write([]byte(challenge))
	return hex.EncodeToString(a.Sum(nil))
}
