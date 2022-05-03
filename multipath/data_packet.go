package multipath

import (
	"bytes"
)

const DATA_PACKET_HEADER_LEN = 12 // header length of data packet (except for payload size)
const DATA_PACKET_PAYLOAD_SIZE = 1024

type DataPacket struct {
	Type      byte
	Length    uint16
	SessionID uint32
	PathID    byte
	SeqNumber uint32
	Payload   []byte
}

func CreateDataPacket(sessionID uint32, pathID int, seq uint32, payload []byte) *DataPacket {
	packet := DataPacket{}
	packet.Type = DATA_PACKET
	packet.Length = uint16(DATA_PACKET_HEADER_LEN + len(payload))
	packet.SessionID = sessionID
	packet.PathID = byte(pathID)
	packet.SeqNumber = seq
	packet.Payload = make([]byte, len(payload))
	copy(packet.Payload, payload)
	return &packet
}

func ParseDataPacket(r *bytes.Reader) (*DataPacket, error) {

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

	pathID, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	seqNumber, err := ReadUint32(r)
	if err != nil {
		return nil, err
	}

	packet := &DataPacket{}
	packet.Type = packetType
	packet.Length = packetLegnth
	packet.SessionID = sessionID
	packet.PathID = pathID
	packet.SeqNumber = seqNumber
	packet.Payload = make([]byte, packetLegnth-DATA_PACKET_HEADER_LEN)
	r.Read(packet.Payload)

	return packet, nil
}

// Writes Data Packet
func (p *DataPacket) Write(b *bytes.Buffer) error {
	b.WriteByte(p.Type)
	WriteUint16(b, uint16(p.Length))
	WriteUint32(b, uint32(p.SessionID))
	b.WriteByte(p.PathID)
	WriteUint32(b, uint32(p.SeqNumber))
	b.Write(p.Payload)

	return nil
}
