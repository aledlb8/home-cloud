package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const storageDir = "storage"

type FileList struct {
	Files []string
}

func main() {
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		os.Mkdir(storageDir, 0755)
	}

	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/download", downloadFile)
	http.HandleFunc("/delete", deleteFile)
	http.HandleFunc("/list", listFilesHandler)
	http.HandleFunc("/mkdir", makeDirectory)
	http.HandleFunc("/", dashboard)

	fmt.Println("Server started at localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	f, err := os.OpenFile(filepath.Join(storageDir, handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, "Failed to open the file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		http.Error(w, "Failed to write the file", http.StatusInternalServerError)
		return
	}
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "File not specified", http.StatusBadRequest)
		return
	}

	f, err := os.Open(filepath.Join(storageDir, filename))
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	_, err = io.Copy(w, f)
	if err != nil {
		http.Error(w, "Failed to read the file", http.StatusInternalServerError)
		return
	}
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "File not specified", http.StatusBadRequest)
		return
	}

	err := os.RemoveAll(filepath.Join(storageDir, filename))
	if err != nil {
		http.Error(w, "Failed to delete the file or directory", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func listFiles() ([]string, error) {
	files, err := ioutil.ReadDir(storageDir)
	if err != nil {
		return nil, err
	}

	fileNames := make([]string, len(files))
	for i, file := range files {
		fileNames[i] = file.Name()
	}

	return fileNames, nil
}

func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	fileNames, err := listFiles()
	if err != nil {
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(FileList{Files: fileNames})
}

func makeDirectory(w http.ResponseWriter, r *http.Request) {
	dirName := r.URL.Query().Get("dir")
	if dirName == "" {
		http.Error(w, "Directory not specified", http.StatusBadRequest)
		return
	}

	err := os.Mkdir(filepath.Join(storageDir, dirName), 0755)
	if err != nil {
		http.Error(w, "Failed to create the directory", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func dashboard(w http.ResponseWriter, r *http.Request) {
    fileNames, err := listFiles()
    if err != nil {
        http.Error(w, "Failed to list files", http.StatusInternalServerError)
        return
    }

    tmpl := template.Must(template.ParseFiles("dashboard.html"))
    err = tmpl.Execute(w, FileList{Files: fileNames})
    if err != nil {
        http.Error(w, "Failed to render the dashboard", http.StatusInternalServerError)
        return
    }
}