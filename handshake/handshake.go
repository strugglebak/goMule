package handshake

import (
	"fmt"
	"io"
)

type Handshake struct {
	ProtocolIdentifier 	string
	InfoHash 						[20]byte
	PeerID							[20]byte
}
// 序列化 handshake 数据成一个 字节数组
func (handshake *Handshake) serialize() []byte {
	// 1. InfoHash 20 个字节
	// 2. PeerID 20 个字节
	// 3. 保留 8 个字节
	// 4. 1 个字节表示整个 handshake 的长度 length
	const offset = 20 + 20 + 8 + 1
	buffer := make([]byte, len(handshake.ProtocolIdentifier)+offset)

	buffer[0] = byte(len(handshake.ProtocolIdentifier))

	index := 1
	index += copy(buffer[index:], handshake.ProtocolIdentifier)
	// 保留 8 个字节
	index += copy(buffer[index:], make([]byte, 8))
	index += copy(buffer[index:], handshake.InfoHash[:])
	index += copy(buffer[index:], handshake.PeerID[:])

	return buffer
}

func buildHandshake(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		ProtocolIdentifier: "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID: peerID,
	}
}

// 从 stream 中解析 handshake
func read(reader io.Reader) (*Handshake, error) {
	// 先解析这个 handshake 有多长
	lengthBuffer := make([]byte, 1)
	_, err := io.ReadFull(reader, lengthBuffer)
	if err != nil {
		return nil, err
	}

	// 根据 handshake 有多长得到 protocolIdentifierLength
	protocolIdentifierLength := int(lengthBuffer[0])
	if protocolIdentifierLength == 0 {
		err = fmt.Errorf("protocol identifier string length cannot be 0")
		return nil, err
	}

	// 解析 handshake 整个消息体
	const offset = 20 + 20 + 8
	handshakeBuffer := make([]byte, protocolIdentifierLength + offset)
	_, err = io.ReadFull(reader, handshakeBuffer)
	if err != nil {
		return nil, err
	}

	// 构建 handshake 数据结构
	var infoHash, peerID [20]byte
	infoHashStartIndex := protocolIdentifierLength + 8
	infoHashEndIndex := protocolIdentifierLength + 8 + 20
	copy(infoHash[:], handshakeBuffer[infoHashStartIndex : infoHashEndIndex])
	copy(peerID[:], handshakeBuffer[infoHashEndIndex:])
	handshake := Handshake {
		ProtocolIdentifier: string(handshakeBuffer[0 : protocolIdentifierLength]),
		InfoHash: infoHash,
		PeerID: peerID,
	}

	return &handshake, nil
}
