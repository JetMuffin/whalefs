package master

import (
	"os"
	"bytes"
	"mime/multipart"
	"path/filepath"
	"net/http"
	"io"
	"testing"
)

var (
	master = NewMaster(config)
)

func newfileUploadRequest(uri string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", uri, body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request, err
}

func TestHTTPServer_Addr(t *testing.T) {
	if master.httpServer.Addr() != "127.0.0.1:8003" {
		t.Error("master http address method error.")
	}
}

func TestHTTPServer_AddrWithScheme(t *testing.T) {
	if master.httpServer.AddrWithScheme() != "http://127.0.0.1:8003" {
		t.Error("master http address with scheme method error.")
	}
}

func TestHTTPServer_ListenAndServe(t *testing.T) {
	go master.httpServer.ListenAndServe()

	request, err := newfileUploadRequest("http://127.0.0.1:8003/upload", "file", "../conf/whale.conf.template")
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{}
	go client.Do(request)
}