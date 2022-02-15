package bitField

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasPiece(t *testing.T) {
	bitField := BitField{0b01010100, 0b01010100}
	outputs := []bool{false, true, false, true, false, true, false, false, false, true, false, true, false, true, false, false, false, false, false, false}
	for i := 0; i < len(outputs); i++ {
		assert.Equal(t, outputs[i], bitField.HasPiece(i))
	}
}

func TestSetPiece(t *testing.T) {
	tests := []struct {
		input BitField
		index int
		output BitField
	} {
		{
			input: BitField{0b01010100, 0b01010100},
			index: 4, // v (set)
			output: BitField{0b01011100, 0b01010100},
		},
		{
			input: BitField{0b01010100, 0b01010100},
			index: 9, // v (noop)
			output: BitField{0b01010100, 0b01010100},
		},
		{
			input: BitField{0b01010100, 0b01010100},
			index: 15, // v (set)
			output: BitField{0b01010100, 0b01010101},
		},
		{
			input: BitField{0b01010100, 0b01010100},
			index: 19, // v (noop)
			output: BitField{0b01010100, 0b01010100},
		},
	}
	for _, test := range tests {
		bitField := test.input
		bitField.SetPiece(test.index)
		assert.Equal(t, test.output, bitField)
	}
}
