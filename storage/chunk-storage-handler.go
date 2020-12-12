package storage

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

// ChunkStorageHandler can store chunks
type ChunkStorageHandler struct {
	stroagePath string
	recorderID  string
	chunkCount  int
	chunkSize   int
	chunkBuffer []byte
	sessionID   string
}

// NewChunkStorageHandler factory
func NewChunkStorageHandler(storagePath, recorderID string, chunkSize int) *ChunkStorageHandler {

	ret := ChunkStorageHandler{
		stroagePath: storagePath,
		recorderID:  recorderID,
		chunkCount:  0,
		chunkSize:   chunkSize,
		chunkBuffer: []byte{},
		sessionID:   strconv.FormatInt(time.Now().UTC().UnixNano(), 10),
	}

	os.Mkdir(ret.stroagePath, 0777)
	return &ret
}

func (csh *ChunkStorageHandler) store(b []byte) {

	csh.chunkBuffer = append(csh.chunkBuffer, b...)

	if len(csh.chunkBuffer) >= csh.chunkSize {
		//domestic-recorder-booth_1613136001080749145_0000000000001149_1613137568493136160.raw
		fileName := fmt.Sprintf("%s_%s_%016d_%s.raw", csh.recorderID, csh.sessionID, csh.chunkCount, strconv.FormatInt(time.Now().UTC().UnixNano(), 10))

		file, err := os.Create(path.Join(csh.stroagePath, fileName))
		defer file.Close()
		if err != nil {
			log.Fatal(err)
		}

		_, err = file.Write(b)
		if err != nil {
			log.Fatal(err)
		}

		csh.chunkBuffer = []byte{}
		csh.chunkCount++
	}

}
