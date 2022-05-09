package main

import (
	"fmt"
	"mp2bs/multipath"
	"net"
	"context"

	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc"

	udp "github.com/docbull/inlab-fabric-udp-proto"
	// protoG "github.com/golang/protobuf/proto"
)

const MSG_SIZE = 4096 // 4KB
const FILE_SIZE = 10485760 // 10MB

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

	msg.grpcListen()
}


func (msg *Message) grpcListen() {

	lis, err := net.Listen("tcp", ":11800")
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	fmt.Println("MPClient: wating client connection...")

	grpcServer := grpc.NewServer(
		grpc.MaxSendMsgSize(FILE_SIZE),
		grpc.MaxRecvMsgSize(FILE_SIZE),
	)
	udp.RegisterUDPServiceServer(grpcServer, msg)
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		fmt.Println(err)
	}
}

func (msg *Message) BlockDataForUDP(ctx context.Context, envelope *udp.Envelope) (*udp.Status, error) {

	fmt.Println("MPClient: Receive Block data from the Peer container")

	msg.Block.Payload = envelope.Payload
	msg.Block.Signature = envelope.Signature
	msg.Block.SecretEnvelope = envelope.SecretEnvelope

	go msg.SendBlock()

	return &udp.Status{Code: udp.StatusCode_Ok}, nil
}

func (msg *Message) SendBlock() {

	addrList := []string{"127.0.0.1:4251", "127.0.0.1:4252"}
	serverAddr := "127.0.0.1:4242"

	// Create Session Manager
	sessionManager := multipath.CreateSessionManager(addrList)

	// Connect to server
	session := sessionManager.Connect(serverAddr)
	fmt.Printf("MPClient: SessionID=%d\n", session.SessionID)

	start, end := 0, 0
	total := 0

	for start < len(msg.Block.Payload) {
		if start + MSG_SIZE < len(msg.Block.Payload) {
			end = start + MSG_SIZE
		} else {
			end = len(msg.Block.Payload)
		}

		/*
		Block := &udp.Envelope{
			Payload:		msg.Block.Payload[start:end],
			Signature:		nil,
			SecretEnvelope:	nil,
		}

		// data marshalling 
		marshalledEnvelope, err := protoG.Marshal(Block)
		if err != nil {
			fmt.Println(err)
			return
		}
		*/

		sendBytes, err := session.Write(msg.Block.Payload[start:end])
		if err != nil {
			panic(err)
		}

		total = total + sendBytes
		fmt.Printf("MPClient: Send %d bytes to server! (total=%d) \n", sendBytes, total)

		start = end
	}

	session.Close()

	fmt.Printf("MPClient: Finish! Send %d bytes to server!!!!!!!!!!1 \n", total)
}
