package main

import (
	"fmt"
	"mp2bs/multipath"
	"os"
)

const fileSize = 5000000 // const fileSize = 116549808

func main() {

	addrList := []string{"127.0.0.1:4242", "127.0.0.1:4243"}

	// Create Session Manager
	sessionManager := multipath.CreateSessionManager(addrList)

	// Accept
	session := sessionManager.Accept()
	fmt.Printf("MPServer: SessionID=%d\n", session.SessionID)

	// File open
	fo, err := os.Create("receive.dat")
	if err != nil {
		panic(err)
	}

	total := 0
	i := 0

	for {
		buf := make([]byte, 4096)
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

		// Write block data into file
		_, err = fo.Write(buf[:recvBytes])
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("MPServer: Receive %d byte from client! \n", total)
}
