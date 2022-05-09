package main

import (
	"fmt"
	"net"
	"context"

	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc"

	udp "github.com/docbull/inlab-fabric-udp-proto"
)

const MSG_SIZE = 4096 // 4KB
const FILE_SIZE = 10485760 // 10MB

var count = 0

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

	lis, err := net.Listen("tcp", ":16220")
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

	fmt.Println("Server: Receive Block data from the MPServer")

	msg.Block.Payload = envelope.Payload
	msg.Block.Signature = envelope.Signature
	msg.Block.SecretEnvelope = envelope.SecretEnvelope

	count += len(msg.Block.Payload)

	fmt.Printf("Received data size: %d // total: %d \n", len(msg.Block.Payload), count)

	return &udp.Status{Code: udp.StatusCode_Ok}, nil
}
