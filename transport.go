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
	c.SetReadDeadline(t)
	n, err := c.TCPConn.Read(b)
	c.SetReadDeadline(time.Time{})
	return n, err
}
func (c *TSocket) Write(b []byte) (int, error) {
	t := time.Now().Add(c.timeout)
	c.SetWriteDeadline(t)
	n, err := c.TCPConn.Write(b)
	c.SetWriteDeadline(time.Time{})
	return n, err
}

var ErrTransportClose = errors.New("链接断开")

func NewTransport(con *net.TCPConn, timeout time.Duration) *Transport {
	return &Transport{con: &TSocket{con, timeout}, readChan: make(chan []byte, 10), writeChan: make(chan []byte, 10)}
}

func (t *Transport) ReadData() *NetPacket {
	s, ok := <-t.readChan
	if !ok {
		fmt.Println("t readData nil")
		return nil
	}

	// 这里可以处理协议相关的解析工作
	// Here you can deal with protocol related parsing work
	p := &NetPacket{Rw: t}
	err := p.Decoder(s)
	if err != nil {
		fmt.Println("t  decoder nil")
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
	fmt.Println("----t writeData begin")

	if t.close {
		fmt.Println("transport end", string(data))
		return ErrTransportClose
	}

	t.writeChan <- data
	fmt.Println("----t writeData over")
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
	fmt.Println("t begintoread")
	r := bufio.NewReader(t.con)
	for {
		// set deadline
		// t.con.SetReadDeadline(30 * time.Second)
		fmt.Println("--------t read begin")

		err = binary.Read(r, binary.LittleEndian, &headLen)
		if err != nil {
			fmt.Println("--------t read head err", err)
			return
		}
		buf := make([]byte, headLen)
		index := 0
		try := 0
		for index < int(headLen) {
			fmt.Println("--------t read full begin", index, headLen)
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
			fmt.Println("--------t read full", index, headLen)
		}
		fmt.Println(" t read over")
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

	for buf := range t.writeChan {
		fmt.Println("write begin")
		if buf == nil {
			fmt.Println("write nil return")
			continue
		}
		headLen = uint64(len(buf))
		err = binary.Write(t.con, binary.LittleEndian, headLen)
		if err != nil {
			fmt.Println("write head err", err)
			return
		}
		fmt.Println("write begin bogy")
		n, err := t.con.Write(buf)
		if err != nil || n != len(buf) {
			fmt.Println("--->transport write err", err)
			return
		}
		// fmt.Println("write begin flush")
		// if err = w.Flush(); err != nil {
		// fmt.Println("--->transport flush err", err)
		// return
		// }
		fmt.Println("write end")
	}
}
