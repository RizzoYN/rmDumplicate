package fileHash

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"os"
	fp "path/filepath"
	"runtime"
	"slices"
	"sort"
	"sync"
)

var wg = sync.WaitGroup{}
var goNum = runtime.NumCPU()

type FilesHash struct {
	Files           []string
	FilesMD5        []string
	FilesSHA256     []string
	FilesDumplicate []bool
	FileSelected    []string
	mixHash         []string
	ErrFlag         bool
}

func NewFilesHash(path string, sameDir, mixMode bool) *FilesHash {
	maps := make(map[string]int)
	c := make(chan struct{}, goNum)
	defer close(c)
	filesHash := new(FilesHash)
	filesHash.ErrFlag = false
	files, ch, fileNums := filesHash.getAllFiles(path)
	filesHash.Files = files
	filesHash.mixHash = make([]string, fileNums)
	filesHash.FilesMD5 = make([]string, fileNums)
	filesHash.FilesSHA256 = make([]string, fileNums)
	res := make(chan [3]string, fileNums)
	defer close(res)
	var fileSelected []string
	for i := 0; i < fileNums; i++ {
		wg.Add(1)
		c <- struct{}{}
		go filesHash.computeFileHash(ch, c, sameDir, mixMode, res)
	}
	wg.Wait()
	for idx := 0; idx < fileNums; idx++ {
		hashs := <-res
		filesHash.mixHash[idx] = hashs[0]
		filesHash.FilesMD5[idx] = hashs[1]
		filesHash.FilesSHA256[idx] = hashs[2]
		_, in := maps[hashs[0]]
		if !in {
			maps[hashs[0]] = 0
		} else {
			maps[hashs[0]] += 1
			fileSelected = append(fileSelected, filesHash.Files[idx])
		}
	}
	filesHash.FileSelected = fileSelected
	close(ch)
	filesHash.sort(fileNums)
	return filesHash
}

func (f *FilesHash) computeFileHash(ch chan string, c chan struct{}, sameDir, mixMode bool, res chan [3]string) {
	const filechunk = 4096 * 1024
	filePath := <-ch
	dir, _ := fp.Split(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		f.ErrFlag = true
	}
	defer file.Close()
	info, _ := file.Stat()
	filesize := info.Size()
	blocks := uint64(math.Ceil(float64(filesize) / float64(filechunk)))
	hashMD5 := md5.New()
	hashSHA256 := sha256.New()
	var hash string
	for i := uint64(0); i < blocks; i++ {
		blocksize := int(math.Min(filechunk, float64(filesize-int64(i*filechunk))))
		buf := make([]byte, blocksize)
		file.Read(buf)
		io.WriteString(hashMD5, string(buf))
		io.WriteString(hashSHA256, string(buf))
	}
	md5Value := fmt.Sprintf("%x", hashMD5.Sum(nil))
	sha256Value := fmt.Sprintf("%x", hashSHA256.Sum(nil))
	if mixMode {
		hash = md5Value + sha256Value
	} else {
		hash = md5Value
	}
	if sameDir {
		hash = hash + dir
	}
	res <- [3]string{hash, md5Value, sha256Value}
	wg.Done()
	<-c
}

func (f *FilesHash) getAllFiles(path string) ([]string, chan string, int) {
	var files []string
	err := fp.Walk(
		path,
		func(p string, info os.FileInfo, err error) error {
			s, e := os.Stat(p)
			if e == nil && !s.IsDir() {
				files = append(files, p)
			}
			return nil
		},
	)
	if err != nil {
		f.ErrFlag = true
	}
	num := len(files)
	ch := make(chan string, num)
	for _, file := range files {
		ch <- file
	}
	return files, ch, num
}

// 排序存在问题
func (f *FilesHash) sort(num int) {
	hashList := make([]string, num)
	list := make([]string, num)
	copy(hashList, f.mixHash)
	copy(list, f.mixHash)
	sort.Strings(hashList)
	maps := make(map[string]int)
	for idx := 0; idx < num; idx++ {
		hash := hashList[idx]
		_, in := maps[hash]
		if !in {
			maps[hash] = 0
		} else {
			maps[hash] += 1
		}
	}
	dFile := make([]string, 0)
	uFile := make([]string, 0)
	dMD5 := make([]string, 0)
	uMD5 := make([]string, 0)
	dSHA256 := make([]string, 0)
	uSHA256 := make([]string, 0)
	boolList := make([]bool, 0)
	for idx := 0; idx < num; idx++ {
		hash := hashList[idx]
		ix := slices.Index(list, hash)
		list[ix] = ""
		fmt.Println(ix)
		fmt.Println(hash)
		fmt.Println(f.Files[ix])
		fmt.Println(f.FilesMD5[ix])
		fmt.Println()
		if maps[hash] > 0 {
			dFile = append(dFile, f.Files[ix])
			dMD5 = append(dMD5, f.FilesMD5[ix])
			dSHA256 = append(dSHA256, f.FilesSHA256[ix])
			boolList = append(boolList, true)
		} else {
			uFile = append(uFile, f.Files[ix])
			uMD5 = append(uMD5, f.FilesMD5[ix])
			uSHA256 = append(uSHA256, f.FilesSHA256[ix])
		}
	}
	f.Files = append(dFile, uFile...)
	f.FilesMD5 = append(dMD5, uMD5...)
	f.FilesSHA256 = append(dSHA256, uSHA256...)
	f.FilesDumplicate = append(boolList, make([]bool, num-len(dFile))...)
}
