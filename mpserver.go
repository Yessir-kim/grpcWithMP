package main

import (
	"fmt"
	"mp2bs/multipath"
	"context"

	udp "github.com/docbull/inlab-fabric-udp-proto"
	"google.golang.org/grpc"
)

const MSG_SIZE = 4096 // 4KB

type Message struct {
	Block	*udp.Envelope
}

func main() {

	msg := &Message{
		Block:			nil,
	}
	msg.Block = &udp.Envelope{
		Payload:		nil,
		Signature:		nil,
		SecretEnvelope:	nil,
	}

	msg.MPServerListen()
}

func (msg *Message) MPServerListen() {

	peerIP := ":16220"
	conn, err := grpc.Dial(peerIP, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	server := udp.NewUDPServiceClient(conn)

	addrList := []string{"127.0.0.1:4242", "127.0.0.1:4243"}

	// Create Session Manager
	sessionManager := multipath.CreateSessionManager(addrList)

	// Accept
	session := sessionManager.Accept()
	fmt.Printf("MPServer: SessionID=%d\n", session.SessionID)

	total := 0
	i := 0

	for {
		buf := make([]byte, MSG_SIZE)
		recvBytes, err := session.Read(buf)
		if err != nil {
			panic(err)
		}
		if recvBytes <= 0 {
			break
		}

		total = total + recvBytes
		fmt.Printf("MPServer: Read len [%d] : n=%d, total=%d \n", i, recvBytes, total)
		i++

		msg.Block.Payload = buf[:recvBytes]

		res, err := server.BlockDataForUDP(context.Background(), msg.Block)
		if err != nil {
			fmt.Println(err)
			return
		}
		if res.Code != udp.StatusCode_Ok {
			fmt.Println("MPServer: Not OK for block transmission:", res.Code)
			return
		} else {
			fmt.Println("MPServer: Received message from Peer:", res)
		}
	}

	fmt.Printf("MPServer: Receive %d byte from MPClient! \n", total)
}


