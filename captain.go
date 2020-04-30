package squad

import (
	"flag"
	"log"
	"net"
)

var captainAddr *string = flag.String("captain_addr", ":11111", "")

type Captain struct {
	Team []*RemoteSquadMember
}

func NewCaptain() *Captain {
	c := new(Captain)
	c.Team = make([]*RemoteSquadMember, 0)
	return c
}

func (c *Captain) CommandAll(command string) {
	for _, sm := range c.Team {
		if sm != nil {
			sm.Sending <- []byte(command)
		}
	}
}

func (c *Captain) Command(command string, teamIndex int) {
	c.Team[teamIndex].Sending <- []byte(command)
}

func (c *Captain) Recruit(size int) error {
	addr, err := net.ResolveTCPAddr("tcp", *captainAddr)
	if err != nil {
		log.Fatalf("Failed to resolve local addr: %s. Error: %s", *captainAddr, err)
	}
	log.Printf("Recruiting squad of %d on %s", size, addr.String())
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("An error occured when listening TCP. Error: %s", err)
	}
	if size == -1 {
		for {
			tcpConn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept a connection. Error: %s", err)
			}
			c.Team = append(c.Team, NewRemoteSquadMember(tcpConn))
		}
	} else {
		for s := 0; s < size; s++ {
			tcpConn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept a connection. Error: %s", err)
			}
			c.Team = append(c.Team, NewRemoteSquadMember(tcpConn))
		}
	}
	return nil
}
