package lbbnet

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	log "github.com/donnie4w/go-logger/logger"
)

var MaxSendPackets = 100

type Transport struct {
	con       *TSocket
	readChan  chan []byte
	writeChan chan []byte
	close     bool
	sync.RWMutex
	sync.WaitGroup
}

type TSocket struct {
	*net.TCPConn
	timeout time.Duration
}

func (c *TSocket) RemoteAddr() string {
	return c.TCPConn.RemoteAddr().String()
}
func (c *TSocket) Read(b []byte) (n int, err error) {
	if c.timeout > time.Nanosecond {
		t := time.Now().Add(c.timeout)
		c.SetReadDeadline(t)
		n, err = c.TCPConn.Read(b)
		c.SetReadDeadline(time.Time{})
	} else {
		n, err = c.TCPConn.Read(b)
	}
	return n, err
}
func (c *TSocket) Write(b []byte) (n int, err error) {
	if c.timeout > time.Nanosecond {
		t := time.Now().Add(c.timeout)
		c.SetWriteDeadline(t)
		n, err = c.TCPConn.Write(b)
		c.SetWriteDeadline(time.Time{})
	} else {
		n, err = c.TCPConn.Write(b)
	}
	return n, err
}

var ErrTransportClose = errors.New("链接断开")

func NewTransport(con *net.TCPConn, timeout time.Duration) *Transport {
	return &Transport{con: &TSocket{con, timeout}, readChan: make(chan []byte, 10), writeChan: make(chan []byte, 10)}
}

func (c *Transport) RemoteAddr() string {
	return c.con.RemoteAddr()
}
func (t *Transport) ReadData() *NetPacket {
	s, ok := <-t.readChan
	if !ok {
		log.Debug("t readData nil")
		return nil
	}

	// 这里可以处理协议相关的解析工作
	// Here you can deal with protocol related parsing work
	p := &NetPacket{Rw: t}
	err := p.Decoder(s)
	if err != nil {
		log.Warn("t  decoder nil")
		return nil
	}
	return p
}

func (t *Transport) Close() {
	t.Lock()
	defer t.Unlock()

	if t.close {
		return
	}
	t.close = true
	t.con.Close()
	close(t.writeChan)
}

var ErrEmptyPacket = errors.New("empty packet")

func (t *Transport) WriteData(data *NetPacket) error {
	t.RLock()
	defer t.RUnlock()
	log.Debug("----t writeData begin")

	if t.close {
		log.Warn("send err,transport close")
		return ErrTransportClose
	}
	if data == nil {
		return ErrEmptyPacket
	}

	t.writeChan <- data.Serialize()
	log.Debug("----t writeData over")
	return nil
}

func (t *Transport) BeginWork() {
	t.Add(2)
	go t.beginToRead()
	go t.beginToWrite()
	t.Wait()
}
func (t *Transport) beginToRead() {
	t.Done()
	defer func() {
		t.Close()
		close(t.readChan)
	}()
	var (
		headLen uint32
		err     error
	)
	// buff size = 4k
	log.Debug("t begintoread")
	r := bufio.NewReader(t.con)
	for {
		err = binary.Read(r, binary.LittleEndian, &headLen)
		if err != nil {
			log.Warn("--------t read head err", err)
			return
		}
		buf := make([]byte, headLen)
		index := 0
		try := 0
		for index < int(headLen) {
			log.Debug("--------t read full begin", index, headLen)
			n, err := io.ReadFull(r, buf[index:])
			if err != nil {
				e, ok := err.(net.Error)
				if !ok || !e.Temporary() || try >= 3 {
					log.Warn("-->transport read err", err)
					return
				}
				try++
			}
			index += n
			log.Debug("--------t read full", index, headLen)
		}
		log.Debug(" t read over")
		t.readChan <- buf
	}
}

func (t *Transport) beginToWrite() {
	t.Done()
	defer func() {
		t.Close()
	}()
	var (
		err error
		n   int
	)

	for buf := range t.writeChan {
		if buf == nil {
			log.Warn("write nil return")
			continue
		}
		n, err = t.con.Write(buf)
		if err != nil || n != len(buf) {
			log.Warn("--->transport write err", err)
			return
		}
	}
}
