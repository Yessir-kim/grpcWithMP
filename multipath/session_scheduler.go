package multipath

const (
	SCHED_USER_WRR = 1 // User-defined weight round robin
	SCHED_NET_WRR  = 2 // Newtwork condition based weight round robin
)

const REMAINING_BYTES_RESET_THRESH = DATA_PACKET_PAYLOAD_SIZE / 8

// Multipath session scheduler for packet transmission
type SessionScheduler struct {
	schedulerType  int
	numPath        int
	weight         []uint32
	remainingBytes []uint32
	currentPath    int
}

func CreateSessionScheduler(schedType int) *SessionScheduler {
	c := SessionScheduler{
		schedulerType:  schedType,
		numPath:        0,
		remainingBytes: make([]uint32, 0),
	}

	// Set weight
	if schedType == SCHED_USER_WRR {
		c.weight = make([]uint32, len(CONFIG_USER_WRR_WEIGHT))
		for i := 0; i < len(c.weight); i++ {
			c.weight[i] = CONFIG_USER_WRR_WEIGHT[i]
		}
	}

	return &c
}

func (c *SessionScheduler) SetNumPath(numPath int) {
	c.numPath = numPath

	if c.numPath > 1 {
		// when the additional path is added,
		// reset remaining bytes of current path
		c.remainingBytes[c.currentPath] = c.weight[c.currentPath] * DATA_PACKET_PAYLOAD_SIZE
	}

	// change current path to new path and set the remainig bytes
	c.currentPath = c.numPath - 1
	remainBytesOfNewPath := c.weight[c.currentPath] * DATA_PACKET_PAYLOAD_SIZE

	c.remainingBytes = append(c.remainingBytes, remainBytesOfNewPath)

	Log("SetNumPath=%d, len remainingBytes=%d", numPath, len(c.remainingBytes))
}

// Weighted Round robin
func (c *SessionScheduler) Scheduling(payloadSize uint32) int {
	pathID := 0

	switch c.schedulerType {
	case SCHED_USER_WRR:
		pathID = c.scheduling_user_wrr(payloadSize)

	case SCHED_NET_WRR:
		pathID = c.scheduling_net_wrr(payloadSize)

	default:
		pathID = c.scheduling_user_wrr(payloadSize)
	}

	return pathID
}

// User-defined weight round robin
func (c *SessionScheduler) scheduling_user_wrr(payloadSize uint32) int {
	selectedPath := c.currentPath

	// update remaing bytes and sent bytes of selected path
	c.remainingBytes[selectedPath] -= payloadSize

	// reset remaining bytes of selected path and change the current path to next path
	if c.remainingBytes[selectedPath] <= REMAINING_BYTES_RESET_THRESH {
		c.remainingBytes[selectedPath] = c.weight[selectedPath] * DATA_PACKET_PAYLOAD_SIZE
		c.currentPath = (c.currentPath + 1) % c.numPath
	}

	return selectedPath
}

// TODO need network information
func (c *SessionScheduler) scheduling_net_wrr(payloadSize uint32) int {
	return 0
}
