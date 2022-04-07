package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/je4/ZipFS/v2/pkg/ZipFS"
)

func (sfs *ServerFS) uploadFile(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(10 << 20)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create file
	dst, err := os.Create(handler.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("dst name: ", dst.Name())
	fmt.Fprintf(w, "Successfully Uploaded zip file\n")
}

type ZipRequest struct {
	Filename string
	Path     string
}

func (sfs *ServerFS) readDir(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var zr ZipRequest
	err := decoder.Decode(&zr)
	if err != nil {
		panic(err)
	}

	zfs, err := ZipFS.NewZipFS(os.DirFS("."), zr.Path)
	if err != nil {
		log.Fatalf("cannot open test zip: %v", err)
	}

	defer zfs.Close()

	fname := zr.Filename
	list, err := zfs.ReadDir(fname)
	if err != nil {
		log.Fatalf("cannot get directories from %s in zip: %v", fname, err)
	}
	fmt.Fprintf(w, "Successfully read zip file second entry: %s\n", list[1].Name())
}

type ServerFS struct {
	fs fs.FS
}

func main() {
	serverFS := ServerFS{fs: os.DirFS(".")}

	zfs, err := ZipFS.NewZipFS(os.DirFS("../testdata"), "photo.zip")
	if err != nil {
		log.Fatalf("cannot open zip: %v", err)
	}

	defer zfs.Close()

	// http.Handle("/", http.FileServer(http.FS(serverFS.fs)))
	http.Handle("/", http.FileServer(http.FS(zfs)))
	http.HandleFunc("/upload", serverFS.uploadFile)
	http.HandleFunc("/read", serverFS.readDir)

	err = http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
