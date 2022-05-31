package gobore

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	conn       net.Conn
	to         string
	localHost  string
	localPort  uint16
	remotePort uint16
	auth       *authenticator
}

func NewClient(localHost string, localPort uint16, to string, port uint16, secret string) (*Client, error) {
	c := &Client{localHost: localHost, localPort: localPort, remotePort: port}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", to, controlPort))
	if err != nil {
		return nil, err
	}

	if secret != "" {
		c.auth = newAuthenticator(secret)
		err = c.auth.clientHandshake(conn)
		if err != nil {
			return nil, err
		}
		time.Sleep(200 * time.Millisecond)
	}

	err = sendJson(conn, clientMessage{Type: clientMessageHello, Hello: port})
	if err != nil {
		return nil, err
	}

	sm := serverMessage{}
	err = recvJsonWithTimeout(conn, &sm)
	if err != nil {
		return nil, err
	}
	if sm.Hello != 0 {
		c.remotePort = sm.Hello
	}

	zap.L().Info("connected to server", zap.Uint16("port", c.remotePort))
	zap.L().Info("listening at ", zap.String("host", c.localHost), zap.Uint16("port", c.localPort))

	c.conn = conn
	return c, nil
}

func (c *Client) Listen() error {
	for {
		sm := serverMessage{}
		err := recvJson(c.conn, &sm)
		if err != nil {
			return err
		}

		zap.L().Sugar().Debugf("Receive serverMessage: %#v", sm)

		switch sm.Type {
		case serverMessageHello:
			zap.L().Warn("unexpected hello")
		case serverMessageChallenge:
			zap.L().Warn("unexpected challenge")
		case serverMessageConnection:
			go func() {
				ctrlAddr := fmt.Sprintf("%s:%d", c.to, controlPort)
				ctrlConn, err := net.Dial("tcp", ctrlAddr)
				if err != nil {
					zap.L().Error("Failed to dail", zap.String("ctrlAddr", ctrlAddr))
					return
				}
				if c.auth != nil {
					err = c.auth.clientHandshake(ctrlConn)
					if err != nil {
						zap.L().Error("Failed to clientHandshake", zap.Error(err))
						return
					}
					time.Sleep(200 * time.Millisecond)
				}

				err = sendJson(ctrlConn, clientMessage{Type: clientMessageAccept, Accept: sm.Connection})
				if err != nil {
					zap.L().Error("Failed to sendJson", zap.Error(err))
					return
				}

				dataAddr := fmt.Sprintf("%s:%d", c.localHost, c.localPort)
				dataConn, err := net.Dial("tcp", dataAddr)
				if err != nil {
					zap.L().Error("Failed to dail", zap.String("dataAddr", dataAddr))
					return
				}

				proxy(ctrlConn, dataConn)
			}()
		case serverMessageError:
			zap.L().Error("server error", zap.String("err", sm.Error))
		}
	}
}
