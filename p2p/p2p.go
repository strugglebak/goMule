package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	progressbar "github.com/schollz/progressbar/v3"

	client "github.com/strugglebak/goMule/client"
	message "github.com/strugglebak/goMule/message"
	peers "github.com/strugglebak/goMule/peers"
)

const MaxRequestBlockSize = 2 << 13
const MaxUnfulfilledRequestBacklog = 5

type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceResult struct {
	Index 	int
	Buffer  []byte
}

func (t *Torrent) StartDownloadWorker(
	peer peers.Peer,
	workQueue chan *pieceWork,
	results chan *pieceResult,
) {
	log.Printf("Handshaking with %s...\n", peer.IP)

	c, err := client.BuildClient(peer, t.InfoHash, t.PeerID)
	if err != nil {
		log.Printf("NETWORK ERROR: Could not handshake with %s. Disconnecting!\n", peer.IP)
		return
	}
	defer c.Conn.Close()

	log.Printf("Completed handshake with %s!\n", peer.IP)

	c.SendUnchoke()
	c.SendInterested()

	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.Index) {
			// 将 piece 放入队列
			workQueue <- pw
			continue
		}

		// 下载 piece
		buffer, err := AttemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw
			return
		}

		// check sum
		err = CheckIntegrity(pw, buffer)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.Index)
			workQueue <- pw
			// 这个时候说明 piece 没下完，要继续下
			continue
		}

		c.SendHave(pw.Index)

		// 下载完成之后将结果放入 results
		results <- &pieceResult{pw.Index, buffer}
	}
}

func (t *Torrent) CalculatePieceBounds(index int) (begin int, end int) {
	// index 指的是下的是第几块 piece
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) CalculatePieceSize(index int) int {
	begin, end := t.CalculatePieceBounds(index)
	return end - begin
}

// 下载整个 file ，数据存储在内存中
func (t *Torrent) Download() ([]byte, error) {
	prompt := "downloading " + t.Name + "..."
	bar := progressbar.Default(100 * 100, prompt)

	log.Printf("Starting download for %s...", t.Name)

	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)
	for index, hash := range t.PieceHashes {
		// 先计算 piece size，然后对队列进行初始化
		length := t.CalculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, length}
	}

	// 开始从 peer 那里下载
	for _, peer := range t.Peers {
		go t.StartDownloadWorker(peer, workQueue, results)
	}

	// 将 results 中的数据给 buffer
	buffer := make([]byte, t.Length)
	donePieces := 0
	prevPercent := float64(0)
	for donePieces < len(t.PieceHashes) {
		response := <-results
		begin, end := t.CalculatePieceBounds(response.Index)
		copy(buffer[begin:end], response.Buffer)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		// 除了 main thread 的其他 worker 数
		// numWorkers := runtime.NumGoroutine() - 1
		bar.Add(int(percent*100 - float64(prevPercent)*100))
		// log.Printf(
		// 	"(%0.2f%%) Downloaded piece #%d from %d peers\n",
		// 	percent,
		// 	response.Index,
		// 	numWorkers,
		// )
		prevPercent = percent
	}
	close(workQueue)

	log.Printf("Completed download for %s!", t.Name)

	return buffer, nil
}

type pieceWork struct {
	Index  int
	Hash   [20]byte
	Length int
}
func AttemptDownloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		Index:  pw.Index,
		Buffer: make([]byte, pw.Length),
		Client: c,
	}

	// 设置 deadline 可以使得未响应的 peers 不去阻塞，因为如果没响应就不用等待传数据了
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	// 函数结束后禁止 deadline
	defer c.Conn.SetDeadline(time.Time{})

	for state.Downloaded < pw.Length {
		// 如果没有阻塞，就发送请求，直到请求的状态是 unfulfilled 的
		if !state.Client.Choked {
			for state.Backlog < MaxUnfulfilledRequestBacklog && state.Requested < pw.Length {
				remainingBlockSize := MaxRequestBlockSize
				// 最后一个块可能要比标准的要小
				if remainingBlockSize > (pw.Length-state.Requested)  {
					// 还剩下这么多
					remainingBlockSize = pw.Length - state.Requested
				}

				// 继续请求剩下的
				err := c.SendRequest(pw.Index, state.Requested, remainingBlockSize)
				if err != nil {
					return nil, err
				}
				state.Backlog++
				state.Requested += remainingBlockSize
			}
		}

		// 请求状态变成 unfulfilled 的了，那么就开始解析响应回来的数据
		// 并更改对应的 message 状态
		err := state.ChangeState()
		if err != nil {
			return nil, err
		}
	}

	return state.Buffer, nil
}

// 检查完整性，即 check sum
func CheckIntegrity(pw *pieceWork, buffer []byte) error {
	hash := sha1.Sum(buffer)
	if !bytes.Equal(hash[:], pw.Hash[:]) {
		return fmt.Errorf("index %d failed integrity check", pw.Index)
	}
	return nil
}

type pieceProgress struct {
	Index      int
	Buffer     []byte
	Client     *client.Client
	Downloaded int // 从 peer 那里下载了多少个块数据
	Requested  int // 请求了多少个 byte 的块数据
	Backlog    int // 请求有没有到最大限制(目前是 5 个)
}
func (state *pieceProgress) ChangeState() error {
	msg, err := state.Client.Read()
	if err != nil {
		return err
	}

	// KeepAlive
	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.MessageUnChoke:
		state.Client.Choked = false
	case message.MessageChoke:
		state.Client.Choked = true

	case message.MessageHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.Client.Bitfield.SetPiece(index)

	case message.MessagePiece:
		n, err := message.ParsePiece(state.Index, state.Buffer, msg)
		if err != nil {
			return err
		}
		state.Downloaded += n
		state.Backlog--
	}

	return nil
}
