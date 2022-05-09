package main

import (
	"fmt"
	"context"
	"os"
	"io"

	"google.golang.org/grpc"
	udp "github.com/docbull/inlab-fabric-udp-proto"
)

const FILE_SIZE = 10485760 // 10MB

func main() {

	grpcServerAddr := "127.0.0.1:11800"

	conn, err := grpc.Dial(grpcServerAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	client := udp.NewUDPServiceClient(conn)

	fi, err := os.Open("mpserver") // need to be changed
	if err != nil  {
		panic(err)
	}

	total := 0

	for {
		buf := make([]byte, FILE_SIZE)
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		} else if err == io.EOF {
			break
		}

		Block := &udp.Envelope{
			Payload:		buf[:n],
			Signature:		nil,
			SecretEnvelope:	nil,
		}

		total = total + n

		res, err := client.BlockDataForUDP(context.Background(), Block)
		if err != nil {
			fmt.Println(err)
			return
		}

		if res.Code != udp.StatusCode_Ok {
			fmt.Println("client: Not OK for block transmission:", res.Code)
			return
		} else {
			fmt.Printf("client: send %d bytes to grpc server (total=%d) \n", n, total)
			fmt.Println("client: Received message from grpc server: ", res)
		}
	}

	fmt.Println("client: Transmission done")
}
