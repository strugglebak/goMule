package bitField

// bit field 代表一个 peer 所拥有的 pieces
type BitField []byte


func (bitField BitField) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(bitField) {
		return false
	}

	return (bitField[byteIndex] >> uint(7-offset)) & 1 != 0
}

func (bitField BitField) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(bitField) {
		return
	}

	bitField[byteIndex] |= (1 << uint(7-offset))
}
