package main

import (
	"log"
	"os"

	torrentFile "github.com/strugglebak/goMule/torrent_file"
)

const Port = 6881

func main() {
	inPath := os.Args[1]
	outPath := os.Args[2]

	tf, err := torrentFile.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}

	err = tf.DownloadAndSaveTorrentFile(outPath, Port)
	if err != nil {
		log.Fatal(err)
	}
}
