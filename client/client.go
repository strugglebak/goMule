package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	bitField "github.com/strugglebak/goMule/bit_field"
	handshake "github.com/strugglebak/goMule/handshake"
	message "github.com/strugglebak/goMule/message"
	peers "github.com/strugglebak/goMule/peers"
)

// Client 是一个 peer 的 TCP 连接
type Client struct {
	Conn			net.Conn
	Choked		bool
	Bitfield 	bitField.BitField
	Peer			peers.Peer
	InfoHash	[20]byte
	PeerID		[20]byte
}

func BuildClient(
	peer peers.Peer,
	infoHash,
	peerID [20]byte,
) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3 * time.Second)
	if err != nil {
		return nil, err
	}

	// 握手
	_, err = CompleteHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// 接收 bitField
	bf, err := ReceiveBitField(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn: conn,
		Choked: true,
		Bitfield: bf,
		Peer: peer,
		InfoHash: infoHash,
		PeerID: peerID,
	}, nil
}

func (client *Client) Read() (*message.Message, error) {
	msg, err := message.Read(client.Conn)
	return msg, err
}

func (client *Client) SendRequest(index, begin, length int) error {
	request := message.FormatMessageRequest(index, begin, length)
	_, err := client.Conn.Write(request.Serialize())
	return err
}

func (client *Client) SendInterested() error {
	msg := message.Message{ID: message.MessageInterested}
	_, err := client.Conn.Write(msg.Serialize())
	return err
}

func (client *Client) SendNotInterested() error {
	msg := message.Message{ID: message.MessageNotInterested}
	_, err := client.Conn.Write(msg.Serialize())
	return err
}

func (client *Client) SendUnchoke() error {
	msg := message.Message{ID: message.MessageUnChoke}
	_, err := client.Conn.Write(msg.Serialize())
	return err
}

func (client *Client) SendHave(index int) error {
	msg := message.FormatMessageHave(index)
	_, err := client.Conn.Write(msg.Serialize())
	return err
}

func CompleteHandshake(
	conn net.Conn,
	infoHash,
	peerID [20]byte,
) (*handshake.Handshake, error) {
	// 设置 deadline 为 3s
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	// 函数结束后禁止 deadline
	defer conn.SetDeadline(time.Time{})

	// 发送请求
	request := handshake.BuildHandshake(infoHash, peerID)
	_, err := conn.Write(request.Serialize())
	if err != nil {
		return nil, err
	}

	// 接收回应
	response, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}

	// 检查 infoHash
	if !bytes.Equal(response.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("expected infoHash %x but go %x", response.InfoHash, infoHash)
	}

	return response, nil
}

func ReceiveBitField(conn net.Conn) (bitField.BitField, error) {
	// 设置 deadline 为 5s
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	// 函数结束后禁止 deadline
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err = fmt.Errorf("expected bitField but got %s", msg)
		return nil, err
	}
	if msg.ID != message.MessageBitfield {
		err = fmt.Errorf("expected bitField but got Message ID %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}
