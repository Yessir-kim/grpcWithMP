package multipath

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	quic "github.com/lucas-clemente/quic-go"
)

// For multipath session
type Session struct {
	SessionID         uint32
	numPath           int
	streamList        []quic.Stream
	listenAddrList    []string
	connectedAddrList []string
	sequenceNumber    uint32
	sentBytes         []uint32
	recvBytes         []uint32
	scheduler         *SessionScheduler
	recvBuffer        *RecvBuffer
	goodbye           bool
}

func CreateSession(sessionID uint32, addrList []string) *Session {
	s := Session{
		SessionID:         sessionID,
		numPath:           0,
		streamList:        make([]quic.Stream, 0),
		listenAddrList:    addrList,
		connectedAddrList: make([]string, 0),
		sequenceNumber:    0,
		sentBytes:         make([]uint32, 0),
		recvBytes:         make([]uint32, 0),
		scheduler:         CreateSessionScheduler(SCHED_USER_WRR),
		recvBuffer:        CreateRecvBuffer(),
		goodbye:           false,
	}

	return &s
}

func (s *Session) Connect(addr string) {
	// Connect to Master listener
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		panic(err)
	}

	// TODO bind my IP?
	ip4 := net.ParseIP("127.0.0.1").To4()
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: ip4, Port: 0})
	if err != nil {
		panic(err)
	}

	// TLS configuration
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"socket-programming"},
	}

	// QUIC Dial
	quicSess, err := quic.Dial(udpConn, udpAddr, addr, tlsConf, nil)
	if err != nil {
		panic(err)
	}

	// QUIC OpenStreamSync
	quicStream, err := quicSess.OpenStreamSync(context.Background())
	if err != nil {
		panic(err)
	}

	Log("Session.Connect(): Connect to %s (%s)", addr, quicSess.RemoteAddr().String())

	// Add a created session into session map
	pathID := s.AddStream(quicStream, quicSess.RemoteAddr().String())

	// Send Hello Packet
	s.sendHelloPacket(pathID)

	// Receive hello ack packet
	s.receiveHelloAckPacket(pathID)

	// Start receiver
	s.StartReceiver(pathID)
}

func (s *Session) AddStream(stream quic.Stream, connectedAddr string) int {
	s.streamList = append(s.streamList, stream)
	s.connectedAddrList = append(s.connectedAddrList, connectedAddr)
	s.numPath++
	s.sentBytes = append(s.sentBytes, 0)
	s.recvBytes = append(s.recvBytes, 0)

	return (s.numPath - 1)
}

// Receive Hello ACK Packet
func (s *Session) receiveHelloAckPacket(pathID int) {
	// Get stream
	stream := s.streamList[pathID]

	buf := make([]byte, PACKET_SIZE)

	// Read packet type and length
	_, err := stream.Read(buf[:5])
	if err != nil {
		panic(err)
	}

	r := bytes.NewReader(buf[:5])
	packetType, _ := r.ReadByte()
	packetLength, _ := ReadUint16(r)

	// Read remaing data
	_, err = stream.Read(buf[5:packetLength]) // Read after field of packet length
	if err != nil {
		panic(err)
	}

	// Parse packet
	reader := bytes.NewReader(buf)
	if packetType == HELLO_ACK_PACKET {
		packet, err := ParseHelloAckPacket(reader)
		if err != nil {
			panic(err)
		}
		s.handleHelloAckPacket(packet)
	} else {
		panic(fmt.Sprintf("Session.receiveHelloAckPacket(): Unknown packet type (%d)", packetType))
	}
}

func (s *Session) StartReceiver(pathID int) {
	// Start receiver
	go s.receiver(pathID)
}

// TODO implement DeleteStream()

// Packet receiver
func (s *Session) receiver(pathID int) {
	// Get stream
	stream := s.streamList[pathID]

	for {
		buf := make([]byte, PACKET_SIZE)

		// Receive packet type and length
		_, err := stream.Read(buf[:5])
		if err != nil {
			panic(err)
		}

		r := bytes.NewReader(buf[:5])
		packetType, _ := r.ReadByte()
		packetLength, _ := ReadUint16(r)

		// Receive remaing data
		_, err = stream.Read(buf[5:packetLength]) // Read after field of packet length
		if err != nil {
			panic(err)
		}

		reader := bytes.NewReader(buf)

		// Packet Handling
		switch packetType {
		// Hello Packet or Hello ACK Packet
		case HELLO_PACKET:
		case HELLO_ACK_PACKET:
			// Error case since hello packet is received when the session created
			panic(fmt.Sprintf("Session.receiver(): Error! PathID=%d, Packet Type=%d !!!!!!!!!!!!!!!!!!!!!\n", pathID, packetType))

		// Data Packet
		case DATA_PACKET:
			packet, err := ParseDataPacket(reader)
			if err != nil {
				panic(err)
			}

			s.recvBytes[pathID] += uint32(packet.Length - DATA_PACKET_HEADER_LEN)

			s.handleDataPacket(packet)

		// Goodbye Packet
		case GOODBYE_PACKET:
			packet, err := ParseGoodbyePacket(reader)
			if err != nil {
				panic(err)
			}
			s.handleGoodbyePacket(packet)

		default:
			panic(fmt.Sprintf("Unknown packet type: %d \n", packetType))
		}
	}
}

// Send Hello Packet
func (s *Session) sendHelloPacket(pathID int) {
	Log("Session.SendHelloPacket(): SessionID=%d", s.SessionID)

	// Create Hello Packet and covert into byte[]
	// Session ID of first hello packet is 0.
	// After first hello packet, session ID is greater than 0 (assigned by server).
	packet := CreateHelloPacket(s.SessionID)
	b := &bytes.Buffer{}
	packet.Write(b)

	// Send bytes of packet
	s.SendPacket(b.Bytes(), pathID)
}

// Send Hello Ack Packet
func (s *Session) SendHelloAckPacket(pathID int) {
	Log("Session.SendHelloAckPacket(): SessionID=%d", s.SessionID)

	// TODO: Set Nic Info
	nicInfos := s.getNicInfo()

	// Create Hello ACK Packet and covert into byte[]
	packet := CreateHelloAckPacket(s.SessionID, nicInfos)
	b := &bytes.Buffer{}
	packet.Write(b)

	// Send bytes of packet
	s.SendPacket(b.Bytes(), pathID)
}

// Send Data Packet
func (s *Session) sendDataPacket(payload []byte, pathID int) {
	Log("Session.sendDataPacket(): SessionID=%d, PathID=%d, Len.Payload=%d", s.SessionID, pathID, len(payload))

	// Create Data Packet and covert into byte[]
	packet := CreateDataPacket(s.SessionID, pathID, s.sequenceNumber, payload)
	b := &bytes.Buffer{}
	packet.Write(b)

	// Send bytes of packet
	s.SendPacket(b.Bytes(), pathID)
}

// Send Goodbye Packet
func (s *Session) sendGoodbyePacket(pathID int) {
	Log("Session.sendGoodbyePacket(): SessionID=%d", s.SessionID)

	packet := CreateGoodbyePacket(s.SessionID)
	b := &bytes.Buffer{}
	packet.Write(b)

	// Send bytes of packet
	s.SendPacket(b.Bytes(), pathID)
}

func (s *Session) SendPacket(packet []byte, pathID int) {
	stream := s.streamList[pathID]
	_, err := stream.Write(packet)
	if err != nil {
		log.Println(err)
	}
}

// Handle Hello Ack Packet
func (s *Session) handleHelloAckPacket(packet *HelloAckPacket) {
	Log("Session.handleHelloAckPacket(): SessionID=%d", packet.SessionID)

	// Set to session ID assigned by server
	if s.SessionID == 0 {
		s.SessionID = packet.SessionID
	}

	// Set numPath for scheduler -> scheduler begins to consider an added path
	s.scheduler.SetNumPath(s.numPath)

	// TODO Now we establish all connections immediately
	// but we have to change to establish connection adaptively during transmission
	for i, nicInfo := range packet.NicInfos {
		nicAddr := string(nicInfo.Addr)
		Log("Session.handleHelloAckPacket(): NicInfo[%d]=%s", i, nicAddr)

		// Find the address not yet connected
		connected := false
		for _, conAddr := range s.connectedAddrList {
			if nicAddr == conAddr {
				connected = true
			}
		}

		// If not yet connected address is found, connect to that address
		if !connected {
			s.Connect(nicAddr)
		}
	}
}

// Handle Data Packet
func (s *Session) handleDataPacket(packet *DataPacket) {
	s.recvBuffer.PushPacket(packet)
}

// Goodbye Packet
func (s *Session) handleGoodbyePacket(packet *GoodbyePacket) {
	// Terminate receiver go routine
	Log("Session.handleGoodbyePacket()")
	s.goodbye = true
}

// Read data
func (s *Session) Read(buf []byte) (int, error) {
	// Check whether recvBuffer is empty
	// TODO modify???
	for s.recvBuffer.IsEmpty() && !s.goodbye {
	}

	readLen := s.recvBuffer.Read(buf)

	return readLen, nil
}

// Send data
func (s *Session) Write(buf []byte) (int, error) {
	start, end := 0, 0
	pathID := 0
	total := 0

	for start < len(buf) {
		// Determine the range of payload
		if start+DATA_PACKET_PAYLOAD_SIZE < len(buf) {
			end = start + DATA_PACKET_PAYLOAD_SIZE
		} else {
			end = len(buf)
		}

		payloadSize := uint32(end - start)
		total += int(payloadSize)

		// Scheduling
		pathID = s.scheduler.Scheduling(payloadSize)

		// Send data packet
		s.sendDataPacket(buf[start:end], pathID)
		s.sentBytes[pathID] += payloadSize
		s.sequenceNumber++

		start = end
	}

	return total, nil
}

// TODO this function should be modified to get NIC information automatically
func (s *Session) getNicInfo() []NicInfo {
	nicInfos := make([]NicInfo, len(s.listenAddrList))
	for i := 0; i < len(s.listenAddrList); i++ {
		nicInfos[i].Type = 0
		nicInfos[i].AddrLen = byte(len(s.listenAddrList[i]))
		nicInfos[i].Addr = []byte(s.listenAddrList[i])
	}
	return nicInfos
}

// TODO
func (s *Session) Close() {
	s.sendGoodbyePacket(0)

	time.Sleep(200 * time.Millisecond)
	for _, stream := range s.streamList {
		stream.Close()
	}
}
