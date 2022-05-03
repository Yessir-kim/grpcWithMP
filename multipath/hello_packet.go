package multipath

import (
	"bytes"
)

const HELLO_PACKET_HEADER_LEN = 7 // header length of hello packet

type HelloPacket struct {
	Type      byte
	Length    uint16
	SessionID uint32
}

func CreateHelloPacket(sessionID uint32) *HelloPacket {
	packet := HelloPacket{}
	packet.Type = HELLO_PACKET
	packet.Length = HELLO_PACKET_HEADER_LEN
	packet.SessionID = sessionID
	return &packet
}

func ParseHelloPacket(r *bytes.Reader) (*HelloPacket, error) {

	packetType, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	packetLegnth, err := ReadUint16(r)
	if err != nil {
		return nil, err
	}

	sessionID, err := ReadUint32(r)
	if err != nil {
		return nil, err
	}

	packet := &HelloPacket{}
	packet.Type = packetType
	packet.Length = packetLegnth
	packet.SessionID = sessionID

	return packet, nil
}

// Writes Hello packet
func (p *HelloPacket) Write(b *bytes.Buffer) error {
	b.WriteByte(p.Type)
	WriteUint16(b, uint16(p.Length))
	WriteUint32(b, uint32(p.SessionID))
	return nil
}
