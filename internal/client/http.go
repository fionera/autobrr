package client

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

type DownloadFileResponse struct {
	Body     *io.ReadCloser
	FileName string
}

type HttpClient struct {
	http *http.Client
}

func NewHttpClient() *HttpClient {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	return &HttpClient{
		http: httpClient,
	}
}

func (c *HttpClient) DownloadFile(url string, opts map[string]string) (*DownloadFileResponse, error) {
	if url == "" {
		return nil, nil
	}

	// create md5 hash of url for tmp file
	hash := md5.Sum([]byte(url))
	hashString := hex.EncodeToString(hash[:])
	tmpFileName := fmt.Sprintf("/tmp/%v", hashString)

	// Create the file
	out, err := os.Create(tmpFileName)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("error creating temp file: %v", tmpFileName)
		return nil, err
	}

	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("error downloading file %v from %v", tmpFileName, url)
		return nil, err
	}
	defer resp.Body.Close()

	// retry logic

	if resp.StatusCode != 200 {
		log.Error().Stack().Err(err).Msgf("error downloading file: %v - bad status: %d", tmpFileName, resp.StatusCode)
		return nil, err
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("error writing downloaded file: %v", tmpFileName)
		return nil, err
	}

	// remove file if fail

	res := DownloadFileResponse{
		Body:     &resp.Body,
		FileName: tmpFileName,
	}

	if res.FileName == "" || res.Body == nil {
		log.Error().Stack().Err(err).Msgf("tmp file error - empty body: %v", url)
		return nil, errors.New("error downloading file, no tmp file")
	}

	log.Debug().Msgf("successfully downloaded file: %v", tmpFileName)

	return &res, nil
}