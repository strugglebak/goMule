package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

const (
	MessageChoke 					messageID = 0	// 阻塞接收
	MessageUnChoke 				messageID = 1 // 不阻塞接收
	MessageInterested 		messageID = 2 // 有兴趣接收数据
	MessageNotInterested 	messageID = 3 // 没有兴趣接受数据
	MessageHave 					messageID = 4 // 发送者已经下载好了一个 piece
	MessageBitfield 			messageID = 5 // 对发送者下载好了的一个 piece 进行 encode
	MessageRequest 				messageID = 6 // 从接收者那里请求一块数据
	MessagePiece 					messageID = 7 // 执行请求，交付一块数据
	MessageCancel 				messageID = 8 // 取消请求
)

type Message struct {
	ID 			messageID
	Payload	[]byte
}
func (message *Message) Name() string {
	if message == nil {
		return "KeepAlive"
	}
	switch message.ID {
	case MessageChoke:
		return "Choke"
	case MessageUnChoke:
		return "UnChoke"
	case MessageInterested:
		return "Interested"
	case MessageNotInterested:
		return "NotInterested"
	case MessageHave:
		return "Have"
	case MessageBitfield:
		return "Bitfield"
	case MessageRequest:
		return "Request"
	case MessagePiece:
		return "Piece"
	case MessageCancel:
		return "Cancel"
	default:
		return fmt.Sprintf("Unknown#%d", message.ID)
	}
}

func (message *Message) String() string {
	if message == nil {
		return message.Name()
	}
	return fmt.Sprintf("%s [%d]", message.Name(), len(message.Payload))
}

// 序列化 message，结果为
// ---------------------------------------
// |length| |message ID| |message Payload|
// ---------------------------------------
//    ↓           ↓               ↓
//   4 byte      1 byte          n byte
// 这样的一个字节数组
func (message *Message) Serialize() []byte {
	if message == nil {
		return make([]byte, 4)
	}

	// length = message ID length + message Payload length
	messageLength := uint32(len(message.Payload)+1)
	// length 本身占 4 个字节
	buffer := make([]byte, messageLength+4)

	// 在 buffer 中设置 length
	binary.BigEndian.PutUint32(buffer[0:4], messageLength)
	// 在 buffer 中设置 message ID
	buffer[4] = byte(message.ID)
	// 在 buffer 中设置 message Payload
	copy(buffer[5:], message.Payload)

	return buffer
}

func Read(reader io.Reader) (*Message, error) {
	// 先解析 length
	lengthBuffer := make([]byte, 4)
	_, err := io.ReadFull(reader, lengthBuffer)
	if err != nil {
		return nil, err
	}

	messageLength := binary.BigEndian.Uint32(lengthBuffer)
	// KeepAlive
	if messageLength == 0 {
		return nil, nil
	}

	messageBuffer := make([]byte, messageLength)
	_, err = io.ReadFull(reader, messageBuffer)
	if err != nil {
		return nil, err
	}

	message := Message {
		ID:	messageID(messageBuffer[0]),
		Payload: messageBuffer[1:],
	}

	return &message, nil
}

func FormatMessageRequest(index, begin, length int) *Message {
	const PayloadLength = 12
	payload := make([]byte, PayloadLength)

	binary.BigEndian.PutUint32(payload[0 : PayloadLength/3], uint32(index))
	binary.BigEndian.PutUint32(payload[PayloadLength/3 : PayloadLength/3*2], uint32(begin))
	binary.BigEndian.PutUint32(payload[PayloadLength/3*2 : PayloadLength], uint32(length))

	return &Message{
		ID: MessageRequest,
		Payload: payload,
	}
}

func FormatMessageHave(index int) *Message {
	const PayloadLength = 4
	payload := make([]byte, PayloadLength)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{
		ID: MessageHave,
		Payload: payload,
	}
}

func ParsePiece(index int, buffer []byte, message *Message) (int, error) {
	if message.ID != MessagePiece {
		return 0, fmt.Errorf("expected PIECE (ID %d), got ID %d", MessagePiece, message.ID)
	}

	if len(message.Payload) < 8 {
		return 0, fmt.Errorf("payload too short. %d < 8", len(message.Payload))
	}

	const PayloadLength = 12
	parsedIndex := int(binary.BigEndian.Uint32(message.Payload[0 : PayloadLength/3]))
	if parsedIndex != index {
		return 0, fmt.Errorf("expected index %d, got %d", index, parsedIndex)
	}
	parsedBegin := int(binary.BigEndian.Uint32(message.Payload[PayloadLength/3 : PayloadLength/3*2]))
	if parsedBegin >= len(buffer) {
		return 0, fmt.Errorf("begin offset too high. %d >= %d", parsedBegin, len(buffer))
	}
	parsedData := message.Payload[PayloadLength/3*2:]
	if parsedBegin+len(parsedData) > len(buffer) {
			return 0, fmt.Errorf(
				"parsed data too long [%d] for offset %d with length %d",
				len(parsedData),
				parsedBegin,
				len(buffer),
			)
	}

	copy(buffer[parsedBegin:], parsedData)

	return len(parsedData), nil
}

func ParseHave(message *Message) (int, error) {
	if message.ID != MessageHave {
		return 0, fmt.Errorf("expected HAVE (ID %d), got ID %d", MessageHave, message.ID)
	}

	if len(message.Payload) != 4 {
		return 0, fmt.Errorf("expected payload length %d. got %d", 4, len(message.Payload))
	}

	index := int(binary.BigEndian.Uint32(message.Payload))

	return index, nil
}
