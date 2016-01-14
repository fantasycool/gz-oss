package main

import (
	"github.com/hanwen/go-fuse/fuse"
	"log"
	"oss"
	"strings"
)

func listDir(name string, ossClient *oss.Client, bucket string) []fuse.DirEntry {
	marker := ""
	result := make([]fuse.DirEntry, 0)
	for {
		if name != "" && name[len(name)-1:] != "/" {
			name = name + "/"
		}
		log.Printf("bucket:%s, marker:%s, name:%s, slash:%si \n", bucket, marker, name, "/")
		rst, err := ossClient.ListBucket(bucket, marker, name, "/")
		log.Printf("Result len is:%d \n", len(rst.Objects))
		if err != nil {
			return nil
		}
		marker = rst.NextMarker
		for _, prefix := range rst.Prefixes {
			arrays := strings.Split(prefix[:len(prefix)-1], "/")
			dirEntry := fuse.DirEntry{Name: arrays[len(arrays)-1], Mode: fuse.S_IFDIR}
			result = append(result, dirEntry)
		}
		for _, obj := range rst.Objects {
			if obj.Key[len(obj.Key)-1:] != "/" {
				arrays := strings.Split(obj.Key, "/")
				//file system only need the last item
				dirEntry := fuse.DirEntry{Name: arrays[len(arrays)-1], Mode: fuse.S_IFREG}
				result = append(result, dirEntry)
			}
		}
		if marker == "" {
			log.Printf("let's break out \n", marker)
			break
		}
	}
	return result
}

// func main() {
// 	cfg := &oss.Config{Endpoint: "oss-cn-hangzhou.aliyuncs.com",
// 		Key: "CnEqcrR0koa8qqwL", Secret: "tYrpf8gccRi8JiW19ki8PxWxyK7ki8"}
// 	ossClient, _ := oss.NewClient(cfg)
// 	fmt.Println(listDir("", ossClient, "frio-tegong"))
// }
