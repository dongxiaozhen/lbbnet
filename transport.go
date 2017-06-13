package lbbnet

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Transport struct {
	con       *TSocket
	readChan  chan []byte
	writeChan chan []byte
	close     bool
	sync.RWMutex
}

type TSocket struct {
	*net.TCPConn
	timeout time.Duration
}

func (c *TSocket) Read(b []byte) (int, error) {
	t := time.Now().Add(c.timeout)
	err := c.SetReadDeadline(t)
	if err != nil {
		return 0, err
	}
	return c.Read(b)
}
func (c *TSocket) Write(b []byte) (int, error) {
	t := time.Now().Add(c.timeout)
	c.SetWriteDeadline(t)
	return c.Write(b)
}

var ErrTransportClose = errors.New("链接断开")

func NewTransport(con *net.TCPConn, timeout time.Duration) *Transport {
	return &Transport{con: &TSocket{con, timeout}, readChan: make(chan []byte, 10), writeChan: make(chan []byte, 10)}
}

func (t *Transport) ReadData() *NetPacket {
	s, ok := <-t.readChan
	if !ok {
		return nil
	}

	// 这里可以处理协议相关的解析工作
	// Here you can deal with protocol related parsing work
	p := &NetPacket{Rw: t}
	err := p.Decoder(s)
	if err != nil {
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

func (t *Transport) WriteData(data []byte) error {
	t.RLock()
	defer t.RUnlock()

	if t.close {
		fmt.Println("transport end", string(data))
		return ErrTransportClose
	}

	t.writeChan <- data
	return nil
}

func (t *Transport) BeginWork() {
	go t.beginToRead()
	go t.beginToWrite()
}
func (t *Transport) beginToRead() {
	defer func() {
		t.Close()
		close(t.readChan)
	}()
	var (
		headLen uint64
		err     error
	)
	// buff size = 4k
	r := bufio.NewReader(t.con)
	for {
		// set deadline
		// t.con.SetReadDeadline(30 * time.Second)

		err = binary.Read(r, binary.LittleEndian, &headLen)
		if err != nil {
			return
		}
		buf := make([]byte, headLen)
		index := 0
		try := 0
		for index < int(headLen) {
			n, err := io.ReadFull(r, buf[index:])
			if err != nil {
				e, ok := err.(net.Error)
				if !ok || !e.Temporary() || try >= 3 {
					fmt.Println("-->transport read err", err)
					return
				}
				try++
			}
			index += n
		}
		t.readChan <- buf
	}
}

func (t *Transport) beginToWrite() {
	defer func() {
		t.Close()
	}()
	var (
		headLen uint64
		err     error
	)
	w := bufio.NewWriter(t.con)
	for buf := range t.writeChan {
		headLen = uint64(len(buf))
		err = binary.Write(w, binary.LittleEndian, headLen)
		if err != nil {
			return
		}
		n, err := w.Write(buf)
		if err != nil || n != len(buf) {
			fmt.Println("--->transport write err", err)
			return
		}
		if err = w.Flush(); err != nil {
			fmt.Println("--->transport flush err", err)
			return
		}

	}
}
