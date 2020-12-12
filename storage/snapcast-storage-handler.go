package storage

import (
	"fmt"
	"os"
)

// SnapcastStorageHandler can provide snapcast with a live feed
type SnapcastStorageHandler struct {
	fifoPath   string
	recorderID string
	chunkCount int
	sessionID  string
	file       *os.File
}

// NewSnapcastStorageHandler factory
func NewSnapcastStorageHandler(fifoPath, recorderID string) *SnapcastStorageHandler {

	f, err := os.OpenFile(fifoPath, os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		fmt.Printf("Cannot open fifo: %v. Snapcast will be disabled", err)
	}

	ret := SnapcastStorageHandler{
		fifoPath:   fifoPath,
		recorderID: recorderID,
		chunkCount: 0,
		file:       f,
	}

	return &ret
}

func (ssh *SnapcastStorageHandler) store(b []byte) {

	if ssh.file != nil {
		written := 0
		for written < len(b) {
			n, err := ssh.file.Write(b)
			if err != nil {
				fmt.Printf("Cannot write into fifo: %v", err)
				return
			}
			written += n
			if written < len(b) {
				fmt.Printf("Written: %d of %d\n", written, len(b))
			}
		}

	}
}
