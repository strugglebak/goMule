package message

type messageID uint8

const (
	MessageChoke 					messageID = 0
	MessageUnChoke 				messageID = 1
	MessageInterested 		messageID = 2
	MessageNotInterested 	messageID = 3
	MessageHave 					messageID = 4
	MessageBitfield 			messageID = 5
	MessageRequest 				messageID = 6
	MessagePiece 					messageID = 7
	MessageCancel 				messageID = 8
)

type Message struct {
	ID 			messageID
	Payload	[]byte
}
