package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/je4/ZipFS/v2/pgk/ZipFS"
)

func upload(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(10 << 20)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("myFile")
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

	zfs, err := NewZipFS(os.DirFS("../.."), dst)
	if err != nil {
		log.Fatalf("cannot open test zip: %v", err)
	}
	defer zfs.Close()

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func main() {
	// mux := http.NewServeMux()
	fsys := http.Dir("../testdata")

	fileServer := http.FileServer(fsys)
	http.Handle("/static", fileServer)
	http.HandleFunc("/upload", uploadFile)

	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
