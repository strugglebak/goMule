package torrent_file

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Announce			string
	InfoHash			[20]byte
	PieceHashes		[][20]byte
	PieceLength		int
	Length				int
	Name					string
}

func Open(filePath string) (TorrentFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return TorrentFile{}, err
	}

	// 等待 Open 函数结束，会自动执行 file.Close()
	defer file.Close()

	bt := bencodeTorrent{}
	// 解析种子文件结构，并将对应的字段写入到 bt 中
	err = bencode.Unmarshal(file, &bt)
	if err != nil {
		return TorrentFile{}, err
	}
	return bt.toTorrentFile()
}

type bencodeInfo struct {
	Pieces				string	`bencode:"pieces"`
	PieceLength		int			`bencode:"pieces length"`
	Length				int			`bencode:"length"`
	Name					string	`bencode:"name"`
}
func (bi *bencodeInfo) generateInfoHash() ([20]byte, error) {
	var buffer bytes.Buffer
	// 将 bencodeInfo encode 并写入到 buffer 中
	err := bencode.Marshal(&buffer, *bi)
	if err != nil {
		return [20]byte{}, err
	}

	// 对 buffer 进行 hash
	infoHash := sha1.Sum(buffer.Bytes())
	return infoHash, nil
}
func (bi *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLength := 20
	buffer := []byte(bi.Pieces)
	// 检查 Pieces 长度是否是 bencodeInfo hash 的长度的倍数
	if len(buffer) % hashLength != 0 {
		err := fmt.Errorf("received malformed pieces of length %d", len(buffer))
		return nil, err
	}
	count := len(buffer) / hashLength
	pieceHashes := make([][20]byte, count)

	// 把 Pieces 这个 hash 字符串以 20 个 byte 为单位拷贝进 pieceHashes 中
	for i := 0; i < count; i++ {
		copy(pieceHashes[i][:], buffer[i*hashLength : (i+1)*hashLength])
	}

	return pieceHashes, nil
}

type bencodeTorrent struct {
	Announce	string			`bencode:"announce"`
	Info			bencodeInfo	`bencode:"info"`
}

func (bt *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := bt.Info.generateInfoHash()
	if err != nil {
		return TorrentFile{}, err
	}
	pieceHashes, err := bt.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}

	torrentFile := TorrentFile {
		Announce: bt.Announce,
		InfoHash: infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bt.Info.PieceLength,
		Length: bt.Info.Length,
		Name: bt.Info.Name,
	}

	return torrentFile, nil
}
