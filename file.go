package main

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"io"
	"log"
	"oss"
	// "strings"
	"sync"
)

type OssFile struct {
	nodefs.File
	bucket    string
	fileName  string
	reader    io.ReadCloser
	off       int64
	ossClient *oss.Client
	lock      *sync.Mutex
	bufFile   *BufFile
	size      int64
}

func NewOssFile(bucket string, fileName string, ossClient *oss.Client, lock *sync.Mutex, size int64, bufFile *BufFile) (ossFile *OssFile) {
	ossFile = &OssFile{bucket: bucket, fileName: fileName, ossClient: ossClient, lock: lock, size: size, bufFile: bufFile}
	ossFile.File = nodefs.NewDefaultFile()
	return ossFile
}

var l sync.Mutex

func (f *OssFile) Read(buf []byte, off int64) (fuse.ReadResult, fuse.Status) {
	log.Printf("Start to call read method!off:%d \n", off)
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.bufFile == nil {
		f.bufFile = &BufFile{off: off, data: nil, size: int(f.size), endOff: int64(0), bucket: f.bucket,
			cacheSize: 1500000, lock: &sync.Mutex{}}
	}
	num, err := readFromBuf(f.bufFile, buf, off, f.size, f.ossClient, f.fileName, 0)

	log.Printf("Has read from Buf %d num bytes \n", num)
	if err != nil {
		return nil, fuse.EINVAL
	} else {
		if num <= len(buf) {
			readResultDatas := fuse.ReadResultData(buf[:num])
			return readResultDatas, fuse.OK
		} else {
			log.Printf("Error!!! num %d,len:%d,off:%d \n", num, len(buf), off)
			return fuse.ReadResultData(buf[:len(buf)]), fuse.OK
		}
	}
}
