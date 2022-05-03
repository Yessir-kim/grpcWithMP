package multipath

// TODO: define config type, read from configuration file

var CONFIG_USER_WRR_WEIGHT = [2]uint32{5, 2}

var verbose_mode = true //TODO

const PACKET_SIZE = 1500

const (
	HELLO_PACKET     = 1
	HELLO_ACK_PACKET = 2
	DATA_PACKET      = 3
	GOODBYE_PACKET   = 4
)
