package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Peer struct {
	IP 		net.IP
	Port	uint16
}
func (p Peer) String() string {
	return net.JoinHostPort(
		p.IP.String(),
		strconv.Itoa(int(p.Port)),
	)
}

// 从 buffer 中解析出 peer IP 和 Port
func Unmarshal(buffer []byte)([]Peer, error) {
	// 4 个 IP, 2 个 Port
	const peerSize = 6
	count := len(buffer) / peerSize

	if len(buffer) % peerSize != 0 {
		err := fmt.Errorf("received malformed peers")
		return nil, err
	}

	peers := make([]Peer, count)
	for i := 0; i < count; i++ {
		offset := i * count
		peers[i].IP = net.IP(buffer[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16(
			[]byte(buffer[offset+4 : offset+6]),
		)
	}

	return peers, nil
}
