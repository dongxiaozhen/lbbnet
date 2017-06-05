package lbbnet

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
)

type Transport struct {
	con       *net.TCPConn
	readChan  chan []byte
	writeChan chan []byte
	close     bool
	sync.RWMutex
}

var ErrTransportClose = errors.New("链接断开")

func NewTransport(con *net.TCPConn) *Transport {
	return &Transport{con: con, readChan: make(chan []byte, 10), writeChan: make(chan []byte, 10)}
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
	r := bufio.NewReader(t.con)
	for {
		err = binary.Read(r, binary.LittleEndian, &headLen)
		if err != nil {
			return
		}
		buf := make([]byte, headLen)
		_, err := r.Read(buf)
		if err != nil {
			fmt.Println("-->transport read err", err)
			return
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
