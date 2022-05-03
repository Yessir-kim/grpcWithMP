package main

import (
	"fmt"
	"io"
	"mp2bs/multipath"
	"os"
)

func main() {

	addrList := []string{"127.0.0.1:4251", "127.0.0.1:4252"}
	serverAddr := "127.0.0.1:4242"

	// Create Session Manager
	sessionManager := multipath.CreateSessionManager(addrList)

	// Connect to server
	session := sessionManager.Connect(serverAddr)
	fmt.Printf("MPClient: SessionID=%d\n", session.SessionID)

	// Open file
	fi, err := os.Open("mpserver")
	if err != nil {
		panic(err)
	}

	total := 0
	sendBytes := 0

	// time.Sleep(3 * time.Second)

	for {
		buf := make([]byte, 4096)
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		} else if err == io.EOF {
			break
		}

		sendBytes, err = session.Write(buf[:n])
		if err != nil {
			panic(err)
		}

		total = total + sendBytes
		fmt.Printf("MPClient: Send %d bytes to server! (total=%d) \n", sendBytes, total)
	}

	//time.Sleep(500 * time.Millisecond)
	session.Close()

	fmt.Printf("MPClient: Finish! Send %d bytes to server!!!!!!!!!!1 \n", total)
}
