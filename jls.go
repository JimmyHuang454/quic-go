package quic

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sagernet/quic-go/internal/protocol"
	"github.com/sagernet/quic-go/internal/utils"
)

const (
	ForwardDefaultTimeout = 5 * time.Minute
)

func (s *connection) CloseJLSForward() {
	if s.JLSForwardCon == nil {
		return
	}
	s.JLSForwardCon.Close()
	s.JLSForwardSend.Close()
	s.closeChan <- closeError{err: errors.New("timeout"), immediate: false, remote: false}
	fmt.Println("Closed JLS forwarding.")
}

func (s *connection) JLSHandshakeError(e error, p receivedPacket) {
	if s.config.UseJLS && !strings.Contains(e.Error(), "JLS") {
		s.JLSIsVaild = true
	}
	if s.config.UseJLS && !s.IsClient() && !s.JLSIsVaild {
		s.receivedPackets <- p
	}
}

func (s *connection) IsTimeout() bool {
	return time.Now().Sub(s.JLSForwardLastAliveTime) > ForwardDefaultTimeout
}

func (s *connection) newJLSForward() error {
	var err error
	s.JLSForwardAddr, err = net.ResolveUDPAddr("udp", s.config.FallbackURL)
	fmt.Println(s.JLSForwardAddr)
	if err != nil {
		return err
	}

	udpConn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return err
	}

	s.JLSForwardCon, err = wrapConn(udpConn) // read
	if err != nil {
		udpConn.Close()
		return err
	}
	s.JLSForwardSend = newSendConn(s.JLSForwardCon, s.JLSForwardAddr, packetInfo{}, utils.DefaultLogger) // send

	go func() {
		for !s.IsTimeout() {
			err := s.JLSForwardSend.SetReadDeadline(time.Now().Add(ForwardDefaultTimeout))
			if err != nil {
				break
			}
			p, err := s.JLSForwardSend.ReadPacket()
			if err != nil {
				fmt.Printf(err.Error())
				continue
			}
			s.conn.Write(p.buffer.Data, 0)
			s.JLSForwardLastAliveTime = time.Now()
		}
		s.CloseJLSForward()
	}()

	return nil
}

func (s *connection) JLSForward(p receivedPacket) (bool, error) {
	s.config.FallbackURL = "www.jsdelivr.com:443"
	if s.config.FallbackURL == "" {
		return true, errors.New("FallbackURL is empty.")
	}

	if s.JLSForwardCon == nil {
		err := s.newJLSForward()
		if err != nil {
			return true, err
		}
	}

	err := s.JLSForwardSend.Write(p.buffer.Data, 0)
	if err != nil {
		return true, err
	}
	s.JLSForwardLastAliveTime = time.Now()
	return false, nil
}

func (s *connection) IsClient() bool {
	return s.perspective != protocol.PerspectiveServer
}

func (s *connection) JLSHandler() {
	if !s.config.UseJLS || s.IsClient() || s.JLSIsVaild || !s.JLSIsChecked {
		return
	}
	s.JLSForwardLastAliveTime = time.Now()
	fmt.Println("Forwarding: " + s.LocalAddr().String())
JLSforward:
	for {
		select {
		case closeErr := <-s.closeChan:
			fmt.Println(closeErr.err)
			break JLSforward
		case p := <-s.receivedPackets:
			isClosed, err := s.JLSForward(p)
			if err != nil || isClosed {
				fmt.Println(err)
				break JLSforward
			}
		}
	}
}
