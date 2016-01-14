package main

import (
	//	"fmt"
	"io"
	// "io/ioutil"
	"log"
	"oss"
	// "strings"
	"sync"
)

//var callNum = 0

type BufFile struct {
	off       int64
	data      []byte
	size      int
	endOff    int64
	bucket    string
	cacheSize int
	lock      *sync.Mutex //read times is bigger than write times
}

//num means start index to fill bytes in buf
func readFromBuf(bufFile *BufFile, buf []byte, off int64, fileSize int64, client *oss.Client,
	fileName string, num int) (int, error) {
	//bufFile data range must equal or cover the buf's range
	if bufFile.data == nil || bufFile.off > off || off >= bufFile.endOff {
		log.Printf("Stop to fill the bullet,off:%d \n", off)
		err := fillBufFile(bufFile, off, fileSize, client, fileName)
		if err != nil {
			log.Printf("fillBufFile failed%s \n", err)
			return -1, err
		}
	}
	//get the start index in bufFile.data
	s_off := off - bufFile.off
	resultNum := 0
	//compute how many bytes we need to fill in this time? it is (len(buf) - num[bytes already filled in])
	//if we want to do only once time, our buf.endOffset must bigger than off + ${how many bytes...}
	log.Printf("(int64(len(buf)):%d,off:%d,int64(num):%d \n", int64(len(buf)), off, int64(num))
	log.Printf("Compare result!:%d >= %d \n", bufFile.endOff, (int64(len(buf)) + off - int64(num)))
	if (bufFile.endOff == fileSize && int(off)+len(buf) >= int(fileSize)) || ((bufFile.endOff > off) && bufFile.endOff >= (int64(len(buf))+off-int64(num))) {
		//we should now compute end index in bufFile.data
		//shouled be filled bytes num:len(buf) - num
		//start index in bufFile.data? off - bufFile.off
		//end index in bufFile.data?off -bufFile.off + len(buf) - num == s_off+len(buf)-num
		log.Printf("Go to once read stage num:%d; reusultNum: %d,off:%d \n", num, resultNum, off)
		return copy(buf[num:], bufFile.data[s_off:]), nil
	} else {
		//when logic comes here, it means that we should do serveral times' read from oss
		//our end off set is bufFile's last data
		log.Printf("Step to recursion process,off:%d;endOff:%d;len(buf):%d \n",
			bufFile.off, bufFile.endOff, len(buf))
		e_off := int(bufFile.endOff - bufFile.off)
		log.Printf("num:%d e_off:%d dataLength:%d,s_off:%d, bufFile.endOff %d \n",
			num, e_off, len(buf), s_off, bufFile.endOff)

		copy(buf[num:], bufFile.data[s_off:e_off])
		newnum := num + e_off - int(s_off)
		resultNum = resultNum + newnum
		bufFile.data = nil
		//the next logic is recursion,we need to get a new buFile.data
		log.Printf("Call readFrom Buf off:%d;endOff:%d;len(buf):%d \n", bufFile.off, bufFile.endOff, len(bufFile.data))
		n, err := readFromBuf(bufFile, buf, bufFile.endOff, fileSize, client, fileName, newnum)
		if err != nil {
			log.Printf("recursion failed %s, n:%d \n", err, n)
			return -1, err
		}
		log.Printf("Return n is %d, resultNum is %d \n", n, resultNum)
		resultNum = resultNum + n
	}
	return resultNum, nil
}

func fillBufFile(bufFile *BufFile, off int64, fileSize int64, client *oss.Client, fileName string) error {
	endOff := off + int64(bufFile.cacheSize)
	if endOff > fileSize {
		endOff = fileSize
	}
	_, reader, err := client.GetObjectRange(bufFile.bucket, fileName, off, endOff-1)
	log.Printf("q %d endoff %d \n", off, endOff-1)
	//callNum = callNum + 1
	//if callNum > 1000000 {
	//	return fmt.Errorf("call num exceed!")
	//}
	data := make([]byte, endOff-off)
	if err != nil {
		log.Printf("make failed!len(data):%d, endoff:%d, off:%d", len(data), endOff, off)
		return err
	}
	num := 0
	for {
		l, err := reader.Read(data[num:])
		if err != nil && err != io.EOF {
			log.Printf("reader.Read failed len(data):%d, num:%d\n", err, len(data), num)
			return err
		}
		num = num + l
		if num == len(data) {
			bufFile.data = data
			bufFile.endOff = endOff
			bufFile.off = off
			bufFile.size = int(endOff - off)
			return nil
		}
	}
}
