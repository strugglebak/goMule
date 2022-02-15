package message

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	tests := []struct {
		input  *Message
		output string
	} {
		{nil, "KeepAlive"},
		{&Message{MessageChoke, []byte{1, 2, 3}}, "Choke [3]"},
		{&Message{MessageUnChoke, []byte{1, 2, 3}}, "UnChoke [3]"},
		{&Message{MessageInterested, []byte{1, 2, 3}}, "Interested [3]"},
		{&Message{MessageNotInterested, []byte{1, 2, 3}}, "NotInterested [3]"},
		{&Message{MessageHave, []byte{1, 2, 3}}, "Have [3]"},
		{&Message{MessageBitfield, []byte{1, 2, 3}}, "Bitfield [3]"},
		{&Message{MessageRequest, []byte{1, 2, 3}}, "Request [3]"},
		{&Message{MessagePiece, []byte{1, 2, 3}}, "Piece [3]"},
		{&Message{MessageCancel, []byte{1, 2, 3}}, "Cancel [3]"},
		{&Message{99, []byte{1, 2, 3}}, "Unknown#99 [3]"},
	}

	for _, test := range tests {
		str := test.input.String()
		assert.Equal(t, test.output, str)
	}
}

func TestSerialize(t *testing.T) {
	tests := map[string]struct {
		input  *Message
		output []byte
	} {
		"serialize message": {
			input:  &Message{ID: MessageHave, Payload: []byte{1, 2, 3, 4}},
			output: []byte{0, 0, 0, 5, 4, 1, 2, 3, 4},
		},
		"serialize keep-alive": {
			input:  nil,
			output: []byte{0, 0, 0, 0},
		},
	}

	for _, test := range tests {
		buffer := test.input.Serialize()
		assert.Equal(t, test.output, buffer)
	}
}

func TestRead(t *testing.T) {
	tests := map[string]struct {
		input  []byte
		output *Message
		fails  bool
	} {
		"parse normal message into struct": {
			input:  []byte{0, 0, 0, 5, 4, 1, 2, 3, 4},
			output: &Message{ID: MessageHave, Payload: []byte{1, 2, 3, 4}},
			fails:  false,
		},
		"parse keep-alive into nil": {
			input:  []byte{0, 0, 0, 0},
			output: nil,
			fails:  false,
		},
		"length too short": {
			input:  []byte{1, 2, 3},
			output: nil,
			fails:  true,
		},
		"buffer too short for length": {
			input:  []byte{0, 0, 0, 5, 4, 1, 2},
			output: nil,
			fails:  true,
		},
	}

	for _, test := range tests {
		reader := bytes.NewReader(test.input)
		message, err := Read(reader)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, message)
	}
}

func TestFormatMessageRequest(t *testing.T) {
	message := FormatMessageRequest(4, 567, 4321)
	expected := &Message {
		ID: MessageRequest,
		Payload: []byte{
			0x00, 0x00, 0x00, 0x04, // Index
			0x00, 0x00, 0x02, 0x37, // Begin
			0x00, 0x00, 0x10, 0xe1, // messageLength
		},
	}
	assert.Equal(t, expected, message)
}

func TestFormatMessageHave(t *testing.T) {
	message := FormatMessageHave(4)
	expected := &Message {
		ID: MessageHave,
		Payload: []byte{0x00, 0x00, 0x00, 0x04},
	}
	assert.Equal(t, expected, message)
}

func TestParsePiece(t *testing.T) {
	tests := map[string]struct {
		inputIndex int
		inputBuf   []byte
		inputMessage   *Message
		outputN    int
		outputBuf  []byte
		fails      bool
	} {
		"parse valid piece": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMessage: &Message{
				ID: MessagePiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index
					0x00, 0x00, 0x00, 0x02, // Begin
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
				},
			},
			outputBuf: []byte{0x00, 0x00, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x00},
			outputN:   6,
			fails:     false,
		},
		"wrong message type": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMessage: &Message{
				ID:      MessageChoke,
				Payload: []byte{},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"payload too short": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMessage: &Message{
				ID: MessagePiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index
					0x00, 0x00, 0x00, // Malformed offset
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"wrong index": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMessage: &Message{
				ID: MessagePiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x06, // Index is 6, not 4
					0x00, 0x00, 0x00, 0x02, // Begin
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"offset too high": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMessage: &Message{
				ID: MessagePiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index is 6, not 4
					0x00, 0x00, 0x00, 0x0c, // Begin is 12 > 10
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, // Block
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
		"offset ok but payload too long": {
			inputIndex: 4,
			inputBuf:   make([]byte, 10),
			inputMessage: &Message{
				ID: MessagePiece,
				Payload: []byte{
					0x00, 0x00, 0x00, 0x04, // Index is 6, not 4
					0x00, 0x00, 0x00, 0x02, // Begin is ok
					// Block is 10 long but begin=2; too long for input buffer
					0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x0a, 0x0b, 0x0c, 0x0d,
				},
			},
			outputBuf: make([]byte, 10),
			outputN:   0,
			fails:     true,
		},
	}

	for _, test := range tests {
		n, err := ParsePiece(test.inputIndex, test.inputBuf, test.inputMessage)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.outputBuf, test.inputBuf)
		assert.Equal(t, test.outputN, n)
	}
}

func TestParseHave(t *testing.T) {
	tests := map[string]struct {
		input  *Message
		output int
		fails  bool
	} {
		"parse valid message": {
			input:  &Message{ID: MessageHave, Payload: []byte{0x00, 0x00, 0x00, 0x04}},
			output: 4,
			fails:  false,
		},
		"wrong message type": {
			input:  &Message{ID: MessagePiece, Payload: []byte{0x00, 0x00, 0x00, 0x04}},
			output: 0,
			fails:  true,
		},
		"payload too short": {
			input:  &Message{ID: MessageHave, Payload: []byte{0x00, 0x00, 0x04}},
			output: 0,
			fails:  true,
		},
		"payload too long": {
			input:  &Message{ID: MessageHave, Payload: []byte{0x00, 0x00, 0x00, 0x00, 0x04}},
			output: 0,
			fails:  true,
		},
	}

	for _, test := range tests {
		index, err := ParseHave(test.input)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, index)
	}
}
