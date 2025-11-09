package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type fileData struct {
	Filename    string
	DownloadURL string
	FullURL     string
}

var allowExtension = map[string]bool{
	".png":  true,
	".pdf":  true,
	".jpg":  true,
	".gif":  true,
	".jpeg": true,
}

var allowContent = map[string]bool{
	"image/png":       true,
	"image/jpeg":      true,
	"image/gif":       true,
	"application/pdf": true,
	"text/plain":      true,
}

var (
	fileMap   = make(map[string]string)
	fileMapMu sync.RWMutex
)

// 확장자 검증
func AllowExtension(filename string) bool {
	extension := filepath.Ext(filename)
	lowerExtension := strings.ToLower(extension)

	if _, ok := allowExtension[lowerExtension]; ok {
		return true
	}
	return false
}

// 파일 존재 여부 확인
func fileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else {
		return false, err
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 최대 64MB 크기
		// 대용량 파일 업로드 방지
		r.Body = http.MaxBytesReader(w, r.Body, 64<<20)

		uploadfile, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer uploadfile.Close()

		// 파일 내용 검증
		var b = make([]byte, 512)
		n, err := uploadfile.Read(b)
		if err != nil && err != io.EOF {
			http.Error(w, "파일 읽기 실패", http.StatusInternalServerError)
			return
		}

		ContentType := http.DetectContentType(b[:n])
		allow := false
		for k := range allowContent {
			if strings.HasPrefix(ContentType, k) {
				allow = true
				break
			}
		}

		if !allow {
			http.Error(w, "허용되지 않는 형식입니다", http.StatusBadRequest)
			return
		}

		uploadfile.Seek(0, io.SeekStart)

		// 확장자 검증
		if !AllowExtension(header.Filename) {
			http.Error(w, fmt.Sprintf("'%s'는 허용되지 않은 확장자입니다.\n 허용된 확장자: .png, .pdf, .gif, .jpeg, jpg", header.Filename), http.StatusBadRequest)
			return
		}

		OgName := header.Filename
		ext := filepath.Ext(OgName)
		Cgfilename := UniqueName() + ext

		// 중복 파일명 방지
		fileMapMu.Lock()
		if _, exists := fileMap[OgName]; exists {
			fileMapMu.Unlock()
			http.Error(w, fmt.Sprintf("'%s'는 이미 업로드된 파일명입니다. 다른 이름으로 업로드해주세요", OgName), http.StatusConflict)
			return
		}

		os.MkdirAll("uploads", 0755)
		file, err := os.Create(filepath.Join("uploads", Cgfilename))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		io.Copy(file, uploadfile)

		fileMap[OgName] = Cgfilename
		fileMapMu.Unlock()

		encodeOgName := url.PathEscape(OgName)
		downURL := fmt.Sprintf("/down/%s", encodeOgName)

		// 공유 url 생성
		scheme := "https"
		host := r.Host
		fullURL := fmt.Sprintf("%s://%s/down/%s", scheme, host, encodeOgName)

		tmpl, err := template.ParseFiles("success.html")
		if err != nil {
			http.Error(w, "템플릿 로드 실패", http.StatusInternalServerError)
			return
		}

		data := fileData{
			Filename:    OgName,
			DownloadURL: downURL,
			FullURL:     fullURL,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "upload.html")
}

func UniqueName() string {
	timestamp := time.Now().UnixNano()
	randomNumber, _ := rand.Int(rand.Reader, big.NewInt(10000))

	return fmt.Sprintf("%d&%d", timestamp, randomNumber)
}

func downHandler(w http.ResponseWriter, r *http.Request) {

	encodeOgName := strings.TrimPrefix(r.URL.Path, "/down/")
	if encodeOgName == "" {
		http.Error(w, "파일명이 없습니다 : ", http.StatusBadRequest)
		return
	}

	OgName, err := url.PathUnescape(encodeOgName)
	if err != nil {
		http.Error(w, "잘못된 파일명입니다", http.StatusBadRequest)
		return
	}

	fileMapMu.RLock()
	CgName, exists := fileMap[OgName]
	fileMapMu.RUnlock()

	if !exists {
		http.Error(w, "파일을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	path := filepath.Join("uploads", CgName)

	// 파일 존재 여부 확인
	exist, err := fileExist(path)
	if err != nil || !exist {
		fmt.Printf("파일 존재 확인 중 오류 발생 : %v\n", err)
		http.Error(w, "파일이 존재하지 않습니다", http.StatusNotFound)
		return
	}

	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "파일 열기 실패 : ", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", OgName))
	w.Header().Set("Content-Type", "application/octet-stream")
	// MIME 타입 검증
	w.Header().Set("X-Content-Type-Options", "nosniff")

	io.Copy(w, file)
}

func main() {
	http.HandleFunc("/", uploadHandler)
	http.HandleFunc("/down/", downHandler)

	http.ListenAndServe(":8080", nil)
}
