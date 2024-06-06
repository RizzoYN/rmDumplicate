package fileHash

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"os"
	fp "path/filepath"
)

type FilesHash struct {
	Files           []string
	FilesMD5        []string
	FilesSHA256     []string
	FilesDumplicate []bool
	FileSelected    []string
}

func NewFilesHash(path string, sameDir bool) *FilesHash {
	maps := make(map[string]int)
	filesHash := new(FilesHash)
	filesHash.Files = filesHash.getAllFiles(path)
	fileNums := len(filesHash.Files)
	hashList := make([]string, fileNums)
	md5List := make([]string, fileNums)
	sha256List := make([]string, fileNums)
	boolList := make([]bool, fileNums)
	var fileSelected []string
	var hash string
	for idx, file := range filesHash.Files {
		md5Value, sha256Value, dir := filesHash.computeFileHash(file)
		if sameDir {
			hash = md5Value + sha256Value + dir
		} else {
			hash = md5Value + sha256Value
		}
		hashList[idx] = hash
		md5List[idx] = md5Value
		sha256List[idx] = sha256Value
		_, in := maps[hash]
		if !in {
			maps[hash] = 0
		} else {
			maps[hash] += 1
			fileSelected = append(fileSelected, file)
		}
	}
	for idx := 0; idx < fileNums; idx++ {
		if maps[hashList[idx]] == 0 {
			boolList[idx] = false
		} else {
			boolList[idx] = true
		}
	}
	filesHash.FilesMD5 = md5List
	filesHash.FilesSHA256 = sha256List
	filesHash.FilesDumplicate = boolList
	filesHash.FileSelected = fileSelected
	return filesHash
}

func (f *FilesHash) computeFileHash(filePath string) (string, string, string) {
	const filechunk = 4096 * 1024
	dir, _ := fp.Split(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	info, _ := file.Stat()
	filesize := info.Size()
	blocks := uint64(math.Ceil(float64(filesize) / float64(filechunk)))
	hashMD5 := md5.New()
	hashSHA256 := sha256.New()
	for i := uint64(0); i < blocks; i++ {
		blocksize := int(math.Min(filechunk, float64(filesize-int64(i*filechunk))))
		buf := make([]byte, blocksize)
		file.Read(buf)
		io.WriteString(hashMD5, string(buf))
		io.WriteString(hashSHA256, string(buf))
	}
	return fmt.Sprintf("%x", hashMD5.Sum(nil)), fmt.Sprintf("%x", hashSHA256.Sum(nil)), dir
}

func (f *FilesHash) getAllFiles(path string) []string {
	var files []string
	err := fp.Walk(path, func(path string, info os.FileInfo, err error) error {
		s, e := os.Stat(path)
		if e == nil && !s.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}
