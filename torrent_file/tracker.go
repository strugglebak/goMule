package torrentFile

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
	peers "github.com/strugglebak/goMule/peers"
)

type bencodeTrackerResponse struct {
	Interval	int			`bencode:"interval"`
	Peers			string	`bencode:"peers"`
}

func (torrentFile *TorrentFile) BuildTrackerURL(
	peerID [20]byte,
	port	 uint16,
) (string, error) {
	baseURL, err := url.Parse(torrentFile.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values {
		"info_hash":		[]string{ string(torrentFile.InfoHash[:]) },
		"peer_id":   	 	[]string{ string(peerID[:]) },
		"port":      	 	[]string{ string(strconv.Itoa(int(port))) },
		"uploaded":  	 	[]string{ "0" },
		"downloaded":  	[]string{ "0" },
		"compact":		 	[]string{ "1" },
		"left":				 	[]string{ strconv.Itoa(torrentFile.Length) },
	}

	baseURL.RawQuery = params.Encode()
	return baseURL.String(), nil
}

func (torrentFile *TorrentFile) RequestPeers(
	peerID [20]byte,
	port	 uint16,
) ([] peers.Peer, error) {
	// 构建 tracker url
	trackerURL, err := torrentFile.BuildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	// 发送 get 请求
	httpClient := &http.Client{ Timeout: 15 * time.Second }
	response, err := httpClient.Get(trackerURL)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	// 把 get 请求回的 response 解析并装载进 trackerResponse
	trackerResponse := bencodeTrackerResponse{}
	err = bencode.Unmarshal(response.Body, &trackerResponse)
	if err != nil {
		return nil, err
	}

	return peers.Unmarshal([]byte(trackerResponse.Peers))
}
