package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	fileMap   = make(map[string]string)
	fileMapMu sync.RWMutex
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		//최대 64MB 크기
		err := r.ParseMultipartForm(64 << 20)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		uploadfile, header, err := r.FormFile("file")
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		defer uploadfile.Close()

		OgName := header.Filename
		ext := filepath.Ext(OgName)
		Cgfilename := UniqueName() + ext

		os.MkdirAll("uploads", 0755)
		file, err := os.Create(filepath.Join("uploads", Cgfilename))
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		defer file.Close()

		io.Copy(file, uploadfile)

		fileMapMu.Lock()
		fileMap[Cgfilename] = OgName
		fileMapMu.Unlock()

		shareURL := fmt.Sprintf("/share/%s", Cgfilename)

		fmt.Fprintf(w, "파일 업로드 성공 : %s\n공유 링크: %s", OgName, shareURL)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "upload.html")
}

func UniqueName() string {
	timestamp := time.Now().UnixNano()
	randomNumber, _ := rand.Int(rand.Reader, big.NewInt(10000))

	return fmt.Sprintf("%d-%d", timestamp, randomNumber)
}

func shareHandler(w http.ResponseWriter, r *http.Request) {

	CgName := strings.TrimPrefix(r.URL.Path, "/share/")

	path := filepath.Join("uploads", CgName)

	file, err := os.Open(path)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", CgName))
	w.Header().Set("Content-Type", "application/octet-stream")

	io.Copy(w, file)
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/share/", shareHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("서버 시작 오류:", err)
	}
}
