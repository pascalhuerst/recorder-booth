package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

// HTTPStorageHandler can upload chungs with http post
type HTTPStorageHandler struct {
	server     string
	recorderID string
	sessionID  string
	chunkCount int
	chunkSize  int
	buffer     bytes.Buffer
}

// NewHTTPStorageHandler factory
func NewHTTPStorageHandler(server, recoderID string, chunkSize int) *HTTPStorageHandler {

	ret := &HTTPStorageHandler{
		server:     server,
		recorderID: recoderID,
		sessionID:  strconv.FormatInt(time.Now().UTC().UnixNano(), 10),
		chunkCount: 0,
		chunkSize:  chunkSize,
		buffer:     bytes.Buffer{},
	}

	return ret
}

func (hus *HTTPStorageHandler) store(b []byte) {

	hus.buffer.Write(b)
	if hus.buffer.Len() >= hus.chunkSize {

		toSend := make([]byte, hus.chunkSize)
		hus.buffer.Read(toSend)
		toSendBuffer := bytes.NewBuffer(toSend)

		var requestBody bytes.Buffer
		multiPartWriter := multipart.NewWriter(&requestBody)

		fileName := fmt.Sprintf("%s_%s_%016d_%s.raw", hus.recorderID, hus.sessionID, hus.chunkCount, strconv.FormatInt(time.Now().UTC().UnixNano(), 10))
		hus.chunkCount++

		fileWriter, err := multiPartWriter.CreateFormFile("raw_audio", fileName)
		if err != nil {
			fmt.Printf("Cannot create multi part file writer: %v\n")
		}

		_, err = io.Copy(fileWriter, toSendBuffer)
		if err != nil {
			fmt.Printf("Cannot write frames into form: %v\n", err)
		}
		multiPartWriter.Close()

		// By now our original request body should have been populated, so let's just use it with our custom request
		req, err := http.NewRequest("POST", hus.server, &requestBody)
		if err != nil {
			fmt.Printf("Cannot issue pos request: %v\n", err)
		}
		req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

		// Do the request
		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			fmt.Printf("Cannot execute request: %v\n", err)
			return
		}

		if response.StatusCode != http.StatusOK {
			fmt.Print("HTTPStorageHandlker: Response was not good!\n")
		}
	}
}
