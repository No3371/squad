package squad

import (
	"log"
	"net"
	"strings"
	"sync"
)

type RemoteSquadMember struct {
	Remote      string
	conn        net.Conn
	Received    chan []byte
	Sending     chan []byte
	closeSignal chan struct{}
}

func (s *RemoteSquadMember) OpenComm(onConnError func(err error), onHandleError func(err error), waitAllClose bool) {
	var wg *sync.WaitGroup
	if waitAllClose {
		wg = new(sync.WaitGroup)
		wg.Add(3)
	}
	go func() {
		buf := make([]byte, 1024)
		log.Printf("Reader of %s opened.", s.Remote)
	r_f:
		for {
			select {
			case <-s.closeSignal:
				break r_f
			default:
			}
			r, err := s.conn.Read(buf)
			if err != nil {
				log.Printf("[ERROR] Failed to read from %s. Error: %s", s.Remote, err)
				if onConnError != nil {
					onConnError(err)
				}
				break r_f
			}
			select {
			case <-s.closeSignal:
				break r_f
			case s.Received <- buf[:r]:
			}
		}
		if waitAllClose {
			wg.Done()
		}
		log.Printf("Reader of %s closing.", s.Remote)
	}()
	go func() {
		log.Printf("Writer of %s opened.", s.Remote)
	w_f:
		for {
			select {
			case <-s.closeSignal:
				break w_f
			default:
			}
			_, err := s.conn.Write(<-s.Sending)
			if err != nil {
				log.Printf("[ERROR] Failed to write bytes to remote.")
				if onConnError != nil {
					onConnError(err)
				}
				break w_f
			}
			select {
			case <-s.closeSignal:
				break w_f
			default:
			}
		}
		if waitAllClose {
			wg.Done()
		}
		log.Printf("Writer of %s closing.", s.Remote)

	}()
	go func() {
		log.Printf("Handler of %s loop opened.", s.Remote)
	h_f:
		for {
			select {
			case <-s.closeSignal:
				break h_f
			case b := <-s.Received:
				splitted := strings.SplitN(string(b), ":", 2)
				if recvHandlers == nil {
					log.Printf("No handlers.")
					continue
				}

				if _, ok := recvHandlers[splitted[0]]; !ok {
					log.Printf("No handler for %s.", splitted)
					continue
				}

				if len(splitted) <= 1 {
					if e := recvHandlers[splitted[0]](s, nil); e != nil {
						if onHandleError != nil {
							onHandleError(e)
						}
					}
				} else {
					if e := recvHandlers[splitted[0]](s, &(splitted[1])); e != nil {
						if onHandleError != nil {
							onHandleError(e)
						}
					}
				}
			}
		}
		if waitAllClose {
			wg.Done()
		}
		log.Printf("Handler of %s loop closing.", s.Remote)
	}()
	if waitAllClose {
		wg.Wait()
	}
}

func (s *RemoteSquadMember) Close() {
	close(s.closeSignal)
}

func ConnectToCaptain(remote string) (s *RemoteSquadMember, err error) {
	conn, err := net.Dial("tcp", remote)
	if err != nil {
		log.Printf("Failed to dial to the designated captain address. Error: %s", err)
		return nil, err
	}
	s = NewRemoteSquadMember(conn)
	return s, nil
}

func NewRemoteSquadMember(conn net.Conn) (s *RemoteSquadMember) {
	s = new(RemoteSquadMember)
	s.conn = conn
	s.Received = make(chan []byte, 10)
	s.Sending = make(chan []byte, 10)
	s.closeSignal = make(chan struct{})
	s.Remote = s.conn.RemoteAddr().String()
	return s
}
