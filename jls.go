package quic

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sagernet/quic-go/internal/protocol"
)

const (
	ForwardDefaultTimeout = 5 * time.Minute
	ReadTimeout           = 5 * time.Second
)

func (s *connection) CloseJLSForward() {
	if s.JLSForwardRaw == nil {
		return
	}
	s.JLSForwardRaw.Close()
	s.closeChan <- closeError{err: errors.New("timeout"), immediate: false, remote: false}
	fmt.Println("Closed JLS forwarding.")
}

func (s *connection) JLSHandshakeError(e error, p receivedPacket) {
	if s.config.UseJLS && !strings.Contains(e.Error(), "JLS") {
		s.JLSIsVaild = true
	}
	if s.config.UseJLS && !s.IsClient() && !s.JLSIsVaild {
		s.JLSForwardLastAliveTime = time.Now()
		s.JLSForward(p)
	}
}

func (s *connection) IsTimeout() bool {
	return time.Now().Sub(s.JLSForwardLastAliveTime) > ForwardDefaultTimeout
}

func (s *connection) JLSPrint(msg any) {
	if s.IsClient() {
		fmt.Printf("Client: ")
	} else {
		fmt.Printf("Server: ")
	}
	fmt.Println(msg)
}

func (s *connection) newJLSForward() error {
	var err error
	s.JLSForwardAddr, err = net.ResolveUDPAddr("udp", s.config.FallbackURL)
	fmt.Println(s.JLSForwardAddr)
	if err != nil {
		fmt.Println(err)
		return err
	}

	s.JLSForwardRaw, err = net.ListenUDP("udp", nil)
	if err != nil {
		return err
	}

	go func() {
		buffer := make([]byte, 2000)
		for !s.IsTimeout() {
			err := s.JLSForwardRaw.SetReadDeadline(time.Now().Add(ForwardDefaultTimeout))
			if err != nil {
				break
			}
			n, _, err := s.JLSForwardRaw.ReadFromUDP(buffer)
			if err != nil {
				fmt.Printf(err.Error())
				continue
			}
			p := buffer[:n]
			s.conn.Write(p, protocol.ByteCount(len(p)))
			s.JLSForwardLastAliveTime = time.Now()
		}
		buffer = nil
		s.CloseJLSForward()
	}()

	return nil
}

// FIXME: connection ID will change after handshake. So Server can not catch all packets to forward.
func (s *connection) JLSForward(p receivedPacket) (bool, error) {
	s.config.FallbackURL = "www.jsdelivr.com:443"
	if s.config.FallbackURL == "" {
		return true, errors.New("FallbackURL is empty.")
	}

	if s.JLSForwardRaw == nil {
		err := s.newJLSForward()
		if err != nil {
			return true, err
		}
	}

	_, err := s.JLSForwardRaw.WriteTo(p.data, s.JLSForwardAddr)
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
