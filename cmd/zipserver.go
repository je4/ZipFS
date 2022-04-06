package main

import (
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
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	zfs, err := ZipFS.NewZipFS(os.DirFS("."), dst.Name())
	if err != nil {
		log.Fatalf("cannot open test zip: %v", err)
	}
	// here is a problem? Error when not commented
	// defer zfs.Close()

	log.Println("zfs: ", zfs)

	fname := "zip"
	list, err := zfs.ReadDir(fname)
	if err != nil {
		log.Fatalf("cannot get directories from %s in zip: %v", fname, err)
	}

	sfs.fs = zfs
	fmt.Fprintf(w, "Successfully Uploaded zip file second entry: %s\n", list[1].Name())
}

func displayZip(sfs ServerFS) http.Handler {
	return http.FileServer(http.FS(sfs.fs))
}

type ServerFS struct {
	fs fs.FS
}

func main() {
	serverFS := ServerFS{fs: os.DirFS(".")}

	http.Handle("/", displayZip(serverFS))
	http.HandleFunc("/upload", serverFS.uploadFile)

	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
