/*
`Pool` creates the connection pool with storing `Transport`
*/
package pool

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"shopee/entry_task/internal"
	"sync"
	"time"
)

// tcp and http communicate through net.conn
// Transport strcut creates a net.conn
type Transport struct {
	conn   net.Conn
	closed bool
}

func NewTransport(conn net.Conn) *Transport {
	return &Transport{conn, false}
}

func (t *Transport) Close() error {
	err := t.conn.Close()
	t.closed = true
	return err
}

func (t *Transport) Send(ori []byte) error {
	dst := make([]byte, 8+len(ori))
	binary.BigEndian.PutUint32(dst[0:4], internal.Prefix)
	binary.BigEndian.PutUint32(dst[4:8], uint32(len(ori)))
	copy(dst[8:], ori)

	// log.Println(dst)

	count := 0
	for count < len(dst) {
		length2, err := t.conn.Write(dst[count:])
		if err != nil {
			return err
		}
		count += length2
	}

	return nil
}

func (t *Transport) Receive() ([]byte, error) {
	prefix := make([]byte, 4)
	_, err := io.ReadFull(t.conn, prefix)
	if err != nil {
		return nil, err
	}
	if binary.BigEndian.Uint32(prefix) != internal.Prefix {
		log.Println("transported prefix is", binary.BigEndian.Uint32(prefix), ", while we wanted", internal.Prefix)
		return nil, errors.New("wrong prefix")
	}

	byteLeng := make([]byte, 4)
	_, err = io.ReadFull(t.conn, byteLeng)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(byteLeng)
	if length == 0 {
		return nil, errors.New("info with length 0")
	}

	rst := make([]byte, int(length))
	_, err = io.ReadFull(t.conn, rst)
	if err != nil {
		return nil, err
	}

	return rst, nil
}

// Pool creates a connection pool with the description of Transport
// Pool is maintained by HTTP server mostly
type Pool struct {
	sync.Mutex
	poolChan chan *Transport
	maxOpen  int
}

// create a new pool
func NewPool(maxOpen, minOpen int) (*Pool, error) {
	if maxOpen <= 0 || minOpen > maxOpen {
		return nil, errors.New("maximum open transports is wrong")
	}
	p := &Pool{maxOpen: maxOpen, poolChan: make(chan *Transport, minOpen)}

	for i := 0; i < minOpen; i++ {
		t, err := p.create()
		if err != nil {
			log.Printf("error in preparing the transport: %v", err)
			continue
		}
		p.poolChan <- t
	}
	return p, nil
}

// close the pool
func (p *Pool) Close() {
	p.Lock()
	defer p.Unlock()

	close(p.poolChan)
	for t := range p.poolChan {
		err := t.Close()
		if err != nil {
			log.Print("error in closing a transport in pool", err)
			continue
		}
	}
}

// call function (with time limits)
func (p *Pool) TimingCall(req []byte) ([]byte, error) {
	timeOutChan := make(chan bool)
	var rst []byte
	var err error

	go func() {
		rst, err = p.call(req)
		if err != nil {
			timeOutChan <- false
		} else {
			timeOutChan <- true
		}
	}()

	select {
	case <-time.After(internal.ConnTimeOut):
		log.Printf("timeout in the request: %v", req)
		return rst, err
	case success := <-timeOutChan:
		if success {
			return rst, nil
		} else {
			return rst, err
		}
	}
}

// call function (with no time limits)
func (p *Pool) call(msg []byte) ([]byte, error) {
	// connect and awaits the release
	t := p.connect()
	log.Println("transport connected to the pool")
	defer p.release(t)

	err := t.Send(msg)
	if err != nil {
		t.Close()
		return nil, err
	}
	log.Println("msg sent")
	rsp, err := t.Receive()
	if err != nil {
		t.Close()
		return nil, err
	}
	log.Println("msg received")
	return rsp, nil
}

// in the pool, create a tcp connection and translate it into a Transport
func (p *Pool) create() (*Transport, error) {
	conn, err := net.DialTimeout("tcp", internal.TCPPort, internal.ConnTimeOut)
	if err != nil {
		return nil, err
	}
	return NewTransport(conn), nil
}

func (p *Pool) connect() *Transport {
	select {
	// if got plenty prepared Transport in pool
	case t := <-p.poolChan:
		return t
	default:
		// if delepted the prepared Transport
		t, err := p.create()
		if err != nil {
			log.Println("error in creating a new transport in the connection")
			return nil
		}
		return t
	}
}

// relese the transport connection after it is closed
func (p *Pool) release(t *Transport) error {
	// if it is already closed, return nil
	if t == nil || t.closed {
		return nil
	}

	select {
	// if the pool could hold this Transport, then re-use it.
	// maintain the original size
	case p.poolChan <- t:
		return nil
	default:
		// otherwise close it.
		err := t.Close()
		return err
	}
}
