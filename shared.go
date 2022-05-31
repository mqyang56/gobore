package gobore

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"time"

	"go.uber.org/zap"
)

const (
	controlPort = 7835

	timeout = 3 * time.Second
)

type clientMessageType string

const (
	clientMessageAuthenticate clientMessageType = "Authenticate"
	clientMessageHello        clientMessageType = "Hello"
	clientMessageAccept       clientMessageType = "Accept"
)

type clientMessage struct {
	Type         clientMessageType `json:"type"`
	Authenticate string            `json:"authenticate,omitempty"`
	Hello        uint16            `json:"hello,omitempty"`
	Accept       string            `json:"accept,omitempty"`
}

type serverMessageType string

const (
	serverMessageChallenge  serverMessageType = "Challenge"
	serverMessageHello      serverMessageType = "Hello"
	serverMessageConnection serverMessageType = "Connection"
	serverMessageError      serverMessageType = "Error"
)

type serverMessage struct {
	Type       serverMessageType `json:"type"`
	Challenge  string            `json:"challenge,omitempty"`
	Hello      uint16            `json:"hello,omitempty"`
	Connection string            `json:"connection,omitempty"`
	Error      string            `json:"error,omitempty"`
}

func proxy(ctlConn, dataConn net.Conn) {
	go func() {
		_, err := io.Copy(ctlConn, dataConn)
		if err != nil {
			zap.L().Error("Failed to Copy", zap.Error(err))
			return
		}
	}()

	_, err := io.Copy(dataConn, ctlConn)
	if err != nil {
		zap.L().Error("Failed to Copy", zap.Error(err))
	}
	return
}

func recvJson(r io.Reader, v interface{}) error {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if err != nil {
			return err
		}

		zap.L().Debug("msg", zap.String("msg", string(buf[:n])))

		err = json.Unmarshal(buf[:n], v)
		if err != nil {
			return err
		}
		return nil
	}
}

func recvJsonWithTimeout(r io.Reader, v interface{}) error {
	done := make(chan error)
	go func() {
		err := recvJson(r, v)
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return errors.New("timed out waiting for initial message")
	}
}

func sendJson(w io.Writer, msg interface{}) (err error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}
