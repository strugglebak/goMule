package torrentFile

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 命令行执行 -update
var update = flag.Bool("update", false, "update .golden.json files")

func TestOpen(t *testing.T) {
	testTorrent, err := Open("test_data/archlinux-2019.12.01-x86_64.iso.torrent")
	require.Nil(t, err)

	goldenJsonPath := "test_data/archlinux-2019.12.01-x86_64.iso.torrent.golden.json"
	if *update {
		serialized, err := json.MarshalIndent(testTorrent, "", "  ")
		require.Nil(t, err)
		ioutil.WriteFile(goldenJsonPath, serialized, 0644)
	}

	expected := TorrentFile{}

	goldenJsonFile, err := ioutil.ReadFile(goldenJsonPath)
	require.Nil(t, err)

	err = json.Unmarshal(goldenJsonFile, &expected)
	require.Nil(t, err)

	assert.Equal(t, expected, testTorrent)
}

func TestToTorrentFile(t *testing.T) {
	tests := map[string] struct {
		input 	*bencodeTorrent
		output	TorrentFile
		fails	bool
	} {
		"correct conversion": {
			input: &bencodeTorrent {
				Announce: "http://bttracker.debian.org:6969/announce",
				Info: bencodeInfo {
					Pieces:      "1234567890abcdefghijabcdefghij1234567890",
					PieceLength: 262144,
					Length:      351272960,
					Name:        "debian-10.2.0-amd64-netinst.iso",
				},
			},
			output: TorrentFile {
				Announce: "http://bttracker.debian.org:6969/announce",
				InfoHash: [20]byte {216, 247, 57, 206, 195, 40, 149, 108, 204, 91, 191, 31, 134, 217, 253, 207, 219, 168, 206, 182},
				PieceHashes: [][20]byte {
					{49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106},
					{97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48},
				},
				PieceLength: 262144,
				Length:      351272960,
				Name:        "debian-10.2.0-amd64-netinst.iso",
			},
			fails: false,
		},
		"not enough bytes in pieces": {
			input: &bencodeTorrent {
				Announce: "http://bttracker.debian.org:6969/announce",
				Info: bencodeInfo {
					Pieces:      "1234567890abcdefghijabcdef",
					PieceLength: 262144,
					Length:      351272960,
					Name:        "debian-10.2.0-amd64-netinst.iso",
				},
			},
			output: TorrentFile {},
			fails:  true,
		},
	}

	for _, test := range tests {
		torrentFile, err := test.input.toTorrentFile()
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, test.output, torrentFile)
	}
}
