package storage

import (
	"bytes"
	"fmt"
	"math"
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

	n, err := hus.buffer.Write(b)
	if err != nil || n != len(b) {
		fmt.Printf("ERROR 1: %v n=%d\n", err, n)
	}

	if hus.buffer.Len() >= hus.chunkSize {

		r := 0
		var toSend []byte
		for i := hus.chunkSize; i != 0; i -= r {
			toSend = make([]byte, int(math.Min(float64(i), float64(hus.chunkSize))))

			r, err = hus.buffer.Read(toSend)
			if err != nil {
				fmt.Printf("Cannot read from byte buffer: %v  len(toSend)=%d hus.buffer.Len=%d\n", err, len(toSend), hus.buffer.Len())
			}
		}

		var requestBody bytes.Buffer
		multiPartWriter := multipart.NewWriter(&requestBody)

		fileName := fmt.Sprintf("%s_%s_%016d_%s.raw", hus.recorderID, hus.sessionID, hus.chunkCount, strconv.FormatInt(time.Now().UTC().UnixNano(), 10))
		hus.chunkCount++

		fileWriter, err := multiPartWriter.CreateFormFile("raw_audio", fileName)
		if err != nil {
			fmt.Printf("Cannot create multi part file writer: %v\n", err)
		}

		n, err := fileWriter.Write(toSend)
		if err != nil {
			fmt.Printf("Cannot write frames into form: %v n=%d\n", err, n)
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
