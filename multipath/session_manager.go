package multipath

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"sync"

	quic "github.com/lucas-clemente/quic-go"
)

// Session Manager
type SessionManager struct {
	mutex          sync.Mutex
	numPath        int
	listenerList   []quic.Listener
	listenAddrList []string
	sessionMap     map[uint32]*Session
	sessionChan    chan *Session
}

func CreateSessionManager(addrList []string) *SessionManager {
	// Create SessionManager
	m := SessionManager{
		numPath:        len(addrList),
		listenerList:   make([]quic.Listener, len(addrList)),
		listenAddrList: addrList,
		sessionMap:     make(map[uint32]*Session),
		sessionChan:    make(chan *Session),
	}

	m.listen()

	return &m
}

func (m *SessionManager) listen() {
	var err error

	// TODO QUIC configuration for enhanced QUIC
	config := quic.Config{}

	// QUIC ListenAddr
	for i, addr := range m.listenAddrList {
		Log("ListenAddr[%d]: %s", i, addr)
		m.listenerList[i], err = quic.ListenAddr(addr, generateTLSConfig(), &config)
		if err != nil {
			panic(err)
		}
	}
}

func (m *SessionManager) Accept() *Session {
	// Start go routines for all listen addresses
	for i := 0; i < m.numPath; i++ {
		go m.accept(i, context.Background())
	}

	// blocking until one session is created
	sess := <-m.sessionChan

	// Add a new session into session map
	m.sessionMap[sess.SessionID] = sess

	return sess
}

func (m *SessionManager) accept(pathID int, ctx context.Context) {

	for {
		// QUIC Accept
		quicSess, err := m.listenerList[pathID].Accept(ctx)
		if err != nil {
			panic(err)
		}

		Log("SessionManger.accept(): PathID=%d, Accepted address=%s",
			pathID, quicSess.RemoteAddr().String())

		// QUIC AcceptStream
		quicStream, err := quicSess.AcceptStream(ctx)
		if err != nil {
			panic(err)
		}

		// Receive a Hello Packet
		sessionID := m.receiveHelloPacket(quicStream)

		m.mutex.Lock()
		var sess *Session
		if sessionID == 0 {
			// Assign a new session ID (first connection)
			sessionID = rand.Uint32()

			// Create a new session
			sess = CreateSession(sessionID, m.listenAddrList)
			m.sessionMap[sessionID] = sess
			Log("SessionManager.accept(): New session is created! (SessionID=%d)", sessionID)
		} else {
			// Get an existing session
			var exists bool
			sess, exists = m.sessionMap[sessionID]
			if !exists {
				panic(fmt.Sprintf("SessionManger.accept(): Received session ID (%d) is not 0 but not exists in the session map!", sessionID))
			} else {
				Log("SessionManager.accept(): New connection is added to existing session! (SessionID=%d)", sessionID)
			}
		}

		// Add a created session into session map
		newPathID := sess.AddStream(quicStream, quicSess.RemoteAddr().String())
		m.mutex.Unlock()

		// Send Hello ACK Packet
		sess.SendHelloAckPacket(newPathID)

		// Start a session receiver
		sess.StartReceiver(newPathID)

		// Send channel for Accept()
		m.sessionChan <- sess
	}
}

// Receive Hello Packet
func (s *SessionManager) receiveHelloPacket(quicStream quic.Stream) uint32 {
	buf := make([]byte, HELLO_PACKET_HEADER_LEN)

	// Read Hello Packet from quic stream
	_, err := quicStream.Read(buf[:HELLO_PACKET_HEADER_LEN])
	if err != nil {
		panic(err)
	}

	// Parse packet type
	r := bytes.NewReader(buf[:HELLO_PACKET_HEADER_LEN])
	packetType, _ := r.ReadByte()

	// Parse packet
	if packetType == HELLO_PACKET {
		reader := bytes.NewReader(buf)
		packet, err := ParseHelloPacket(reader)
		if err != nil {
			panic(err)
		}
		Log("SessionManager.receiveHelloPacket(): SessionID=%d", packet.SessionID)

		return packet.SessionID
	} else {
		panic(fmt.Sprintf("SessionManager.receiveHelloPacket(): Unknown initial packet type (%d)", packetType))
	}
}

// Connect
func (m *SessionManager) Connect(addr string) *Session {
	// Create Session
	sess := CreateSession(0, m.listenAddrList)

	sess.Connect(addr)

	return sess
}
