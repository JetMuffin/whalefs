package master

import (
	"net/http"
	"html/template"
	log "github.com/Sirupsen/logrus"
	"strconv"
	. "github.com/JetMuffin/whalefs/types"
	"io/ioutil"
)

type HTTPServer struct {
	Host 	  	string
	Port 	  	int
	blockManager 	*BlockManager
	nodeManager     *NodeManager
}

func (server *HTTPServer) Addr() string {
	return server.Host + ":" + strconv.Itoa(server.Port)
}

func (server *HTTPServer) AddrWithScheme() string {
	return "http://" + server.Addr()
}

func (server *HTTPServer) upload(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "GET" {
		t, err := template.ParseFiles("static/upload.html")
		if err != nil {
			log.Errorf("Unable to render templates: %v", err)
			return
		}

		data := struct {
			Files []*File
		}{
			Files: server.blockManager.ListFile(),
		}
		t.Execute(w, data)
	} else {
		file, header, err := r.FormFile("file")
		if err != nil {
			log.Errorf("Parse form file error: %v", err)
			return
		}
		defer file.Close()

		if err != nil {
			log.Errorf("Cannot read bytes from uploaded file: %v", err)
			return
		}

		bytes, err := ioutil.ReadAll(file)
		fileMeta := NewFile(header.Filename, int64(len(bytes)))
		server.blockManager.AddFile(fileMeta)
		blob := &Blob{
			FileID: fileMeta.ID,
			Name: header.Filename,
			Length: int64(len(bytes)),
			Content: bytes,
		}
		server.blockManager.blobQueue <- blob

		http.Redirect(w, r, "/upload", 301)
	}
}

func (server *HTTPServer) nodes(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/nodes.html")
	if err != nil {
		log.Errorf("Unable to render templates: %v", err)
		return
	}

	data := struct {
		Nodes []*Node
	}{
		Nodes: server.nodeManager.ListNode(),
	}
	t.Execute(w, data)
}

func (server *HTTPServer) ListenAndServe()  {
	http.HandleFunc("/upload", server.upload)
	http.HandleFunc("/nodes", server.nodes)
	log.WithFields(log.Fields{"host": server.Host, "port": server.Port}).Info("HTTP Server start listening.")

	go http.ListenAndServe(server.Addr(), nil)
}

func NewHTTPServer(host string, port int, blockManager *BlockManager, nodeManager *NodeManager) *HTTPServer {
	return &HTTPServer{
		Host: host,
		Port: port,
		blockManager: blockManager,
		nodeManager: nodeManager,
	}
}