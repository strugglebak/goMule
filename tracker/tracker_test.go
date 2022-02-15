package tracker

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	peers "github.com/strugglebak/goMule/peers"
)

func TestBuildTrackerURL(t *testing.T) {
	torrentFile := TorrentFile {
		Announce: "http://bttracker.debian.org:6969/announce",
		InfoHash: [20]byte{ 183, 161, 12, 130, 76, 207, 59, 26, 19, 131, 12, 89, 56, 129, 45, 111, 83, 71, 10, 68 },
		PieceHashes: [][20]byte{
			{ 49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106 },
			{ 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48 },
		},
		PieceLength: 262144,
		Length:      351272960,
		Name:        "debian-10.2.0-amd64-netinst.iso",
	}

	peerID := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	const port uint16 = 6882
	url, err := torrentFile.buildTrackerURL(peerID, port)
	expected := "http://bttracker.debian.org:6969/announce?compact=1&downloaded=0&info_hash=%B7%A1%0C%82L%CF%3B%1A%13%83%0CY8%81-oSG%0AD&left=351272960&peer_id=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&port=6882&uploaded=0"
	assert.Nil(t, err)
	assert.Equal(t, url, expected)
}

func TestRequestPeers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []byte(
			"d" +
				"8:interval" + "i900e" +
				"5:peers" + "12:" +
				string([]byte{
					192, 0, 2, 123, 0x1A, 0xE1, // 0x1AE1 = 6881
					127, 0, 0, 1, 0x1A, 0xE9, // 0x1AE9 = 6889
				}) + "e")
		w.Write(response)
	}))

	defer server.Close()

	torrentFile := TorrentFile {
		Announce: server.URL,
		InfoHash: [20]byte{ 183, 161, 12, 130, 76, 207, 59, 26, 19, 131, 12, 89, 56, 129, 45, 111, 83, 71, 10, 68 },
		PieceHashes: [][20]byte{
			{ 49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106 },
			{ 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48 },
		},
		PieceLength: 262144,
		Length:      351272960,
		Name:        "debian-10.2.0-amd64-netinst.iso",
	}
	peerID := [20]byte{ 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20 }
	const port uint16 = 6882
	expected := []peers.Peer{
		{IP: net.IP{ 192, 0, 2, 123 }, Port: 6881 },
		{IP: net.IP{ 127, 0, 0, 1 }, Port: 6889 },
	}
	p, err := torrentFile.requestPeers(peerID, port)
	assert.Nil(t, err)
	assert.Equal(t, expected, p)
}
