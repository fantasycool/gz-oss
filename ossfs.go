package main

//Now only implements Oss Read file operation

import (
	"flag"
	"fmt"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"log"
	// "os"
	"oss"
	"sync"
	// "strings"
	// "sys"
	// "time"
)

type OssFs struct {
	pathfs.FileSystem
	ossClient *oss.Client
	bucket    string
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}

func (me *OssFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	log.Printf("name:%s", name)
	if name == "/" || name == "\\" || name == "" {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}
	if len(name) > 0 {
		log.Printf("Get attr name %s \n", name)
		objectInfo, err := me.ossClient.GetObjectInfo(me.bucket, name)
		if err != nil {
			log.Printf("%s \n", err)
			return &fuse.Attr{
				Mode: fuse.S_IFDIR | 0777,
				Size: uint64(0),
			}, fuse.OK
		}
		log.Printf("Get attr bucket %s, name %s  \n", me.bucket, name)
		log.Printf("%s's size is %d", name, objectInfo.Size)
		fmt.Printf("This file's size is:%d \n", objectInfo.Size)
		if objectInfo.Size == 0 {
			return &fuse.Attr{
				Mode: fuse.S_IFDIR | 0777,
				Size: uint64(0),
			}, fuse.OK
		} else {
			return &fuse.Attr{
				Mode: fuse.S_IFREG | 0777,
				Size: uint64(objectInfo.Size),
			}, fuse.OK
		}

	}
	log.Printf("Get Attr return nil!%s \n", name)
	return nil, fuse.ENOENT
}

func (me *OssFs) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	dirEntrys := listDir(name, me.ossClient, me.bucket)
	if dirEntrys != nil {
		log.Printf("dirEntrys length:%d name:%s direntrys:%s\n", len(dirEntrys), name, dirEntrys)

		return dirEntrys, fuse.OK
	}
	return nil, fuse.EIO
}

var fileMap = map[uint32]nodefs.File{}

func (me *OssFs) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	// log.Println("Start to call Open method!name:", name, "flags:", flags)
	if x, ok := fileMap[flags]; !ok {
		objectInfo, err := me.ossClient.GetObjectInfo(me.bucket, name)
		log.Printf("%s's size is %d", name, objectInfo.Size)
		if err != nil {
			return nil, fuse.ENOENT
		}
		fmt.Printf("This file's size is:%d \n", objectInfo.Size)
		log.Printf("Start to call Open method.name:%s,flags:%s \n", name, flags, objectInfo.Size)
		file = NewOssFile(me.bucket, name, me.ossClient, new(sync.Mutex), objectInfo.Size, nil)
		return file, fuse.OK
	} else {
		return x, fuse.OK
	}
}

func main() {
	help := flag.String("h", "", "help document")
	bucket := flag.String("b", "frio-tegong", "set oss bucket to use")
	mountPath := flag.String("p", "", "mount path")
	endPoint := flag.String("e", "oss-cn-hangzhou.aliyuncs.com", "oss endpoint to use")
	key := flag.String("k", "", "oss key to use")
	secret := flag.String("s", "", "oss secret to use")

	flag.Parse()
	if len(*help) > 0 {
		fmt.Println(flag.Args())
		return
	}
	if len(*mountPath) == 0 {
		fmt.Errorf("mount path must not be null", "")
		return
	}

	cfg := &oss.Config{Endpoint: *endPoint, Key: *key, Secret: *secret}
	ossClient, err := oss.NewClient(cfg)
	check(err)

	nfs := pathfs.NewPathNodeFs(&OssFs{FileSystem: pathfs.NewDefaultFileSystem(),
		bucket: *bucket, ossClient: ossClient}, nil)
	server, _, err := nodefs.MountRoot(*mountPath, nfs.Root(), nil)
	check(err)
	server.Serve()
}
