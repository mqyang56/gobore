package gobore

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Server struct {
	minPort uint16

	auth *authenticator

	dataConns map[string]net.Conn
	guard     sync.Mutex

	availablePorts chan uint16
}

func NewServer(minPort uint16, secret string, ports []uint16) *Server {
	s := &Server{minPort: minPort}
	s.dataConns = map[string]net.Conn{}
	if secret != "" {
		s.auth = newAuthenticator(secret)
	}

	s.availablePorts = func() chan uint16 {
		ch := make(chan uint16, len(ports))
		for _, port := range ports {
			ch <- port
		}
		return ch
	}()
	return s
}

func (s *Server) Listen() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", controlPort))
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		id := uuid.New().String()

		go func() {
			err = s.handleConn(conn)
			if err != nil {
				zap.L().Warn("connection exited with error", zap.String("id", id), zap.Error(err))
				return
			}

			zap.L().Info("connection exited", zap.String("id", id))
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) error {
	if s.auth != nil {
		err := s.auth.serverHandshake(conn)
		if err != nil {
			err = sendJson(conn, &serverMessage{Type: serverMessageError, Error: err.Error()})
			if err != nil {
				return err
			}
		}
	}

	cm := clientMessage{}
	err := recvJsonWithTimeout(conn, &cm)
	if err != nil {
		return err
	}

	zap.L().Sugar().Debugf("Receive clientMessage: %#v", cm)

	switch cm.Type {
	case clientMessageAuthenticate:
		return errors.New("unexpected authenticate")
	case clientMessageAccept:
		zap.L().Info("forward connection")
		s.guard.Lock()
		dataConn, ok := s.dataConns[cm.Accept]
		s.guard.Unlock()
		if !ok {
			return fmt.Errorf("missing connection %s", cm.Accept)
		}

		proxy(conn, dataConn)
	case clientMessageHello:
		if cm.Hello != 0 && cm.Hello < s.minPort {
			return fmt.Errorf("client port number %d too low", cm.Hello)
		}

		select {
		case cm.Hello = <-s.availablePorts:
			defer func() {
				s.availablePorts <- cm.Hello
			}()
		default:
			return fmt.Errorf("no available port in server")
		}

		zap.L().Info("new client", zap.Uint16("port", cm.Hello))

		l, err := net.Listen("tcp", fmt.Sprintf(":%d", cm.Hello))
		if err != nil {
			return err
		}
		defer func() {
			zap.L().Info("close listener", zap.String("addr", l.Addr().String()))
			_ = sendJson(conn, serverMessage{
				Type:  serverMessageError,
				Error: fmt.Sprintf("Closed connection %d", l.Addr().(*net.TCPAddr).Port),
			})

			_ = l.Close()
		}()

		err = sendJson(conn, serverMessage{Type: serverMessageHello, Hello: uint16(l.Addr().(*net.TCPAddr).Port)})
		if err != nil {
			zap.L().Error("Failed to sendJson", zap.Error(err))
			return err
		}

		var count int
		go func() {
			for {
				dataConn, err := l.Accept()
				if err != nil {
					return
				}

				zap.L().Info("new connection", zap.String("addr", dataConn.RemoteAddr().String()))
				count = 0

				id := uuid.New().String()
				s.guard.Lock()
				s.dataConns[id] = dataConn
				s.guard.Unlock()
				err = sendJson(conn, serverMessage{Type: serverMessageConnection, Connection: id})
				if err != nil {
					zap.L().Error("Failed to sendJson", zap.String("id", id))
				}
				go func() {
					time.Sleep(10 * time.Second)
					s.guard.Lock()
					delete(s.dataConns, id)
					s.guard.Unlock()
				}()
			}
		}()

		t := time.NewTicker(30 * time.Minute)
		defer t.Stop()
		for range t.C {
			count++
			if count >= 1 {
				return nil
			}
		}
	}
	return nil
}
