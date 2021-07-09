package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

type Download struct {
	concurrency int //并发数
}

func NewDownload(concurrency int) *Download  {
	return &Download{
		concurrency : concurrency,
	}
}

func (d*Download) download(url string) error {
	//判断是否能到达
	resp,err:=http.Head(url)
	if err!=nil {

	}
	fileName:=path.Base(url)
	//如果能并发下载
	if resp.Header.Get("Accept-Ranges")=="bytes"&&resp.StatusCode==http.StatusOK {
	return d.Multidownload(url,int(resp.ContentLength),fileName)
	}
	return d.Singledownload(url,int(resp.ContentLength),fileName)

}

func (d*Download) Multidownload(url string,len int,fileName string) error{
	//计算每块的大小
	size:=len/d.concurrency
	
	//创建部分文件的存放目录
	partDir:=d.getPartDir(fileName)
	os.Mkdir(partDir,0777)
	var wg sync.WaitGroup
	wg.Add(d.concurrency)

	start:=0

	for i:=0;i<d.concurrency;i++ {
		go func(start ,i int) {
			defer wg.Done()
			end:=start+size
			if(i==d.concurrency-1){

				//实际下载
			end=len
			}

			d.downloadPartial(start,end,i,fileName,url)
		}(start,i)
		start+=size+1
	}
	wg.Wait()
	d.merge(fileName)
	return nil
	
}
func (d*Download) merge(fileName string) error{
	destFile,err:=os.OpenFile(fileName,os.O_CREATE|os.O_WRONLY,0666)
	if err!=nil{

	}
	defer destFile.Close()
	for i:=0;i<d.concurrency;i++ {
		partFileName:=d.getPartFilename(fileName,i)
		partFile,err:=os.Open(partFileName)
		if err!=nil{

		}
		io.Copy(destFile,partFile)
		partFile.Close()
		os.Remove(partFileName)
	}
	return nil
}
func (d*Download) Singledownload(url string,len int,fileName string) error{
	resp,err:=http.Get(url)
	defer resp.Body.Close()//为什么需要关上resp.body
	if err!=nil {

	}
	file,err:=os.OpenFile(fileName,os.O_CREATE|os.O_WRONLY, 0666)
	if err!=nil {

	}
	defer file.Close()
	buf:=make([]byte,32*1024)
	_,err=io.CopyBuffer(io.MultiWriter(file),resp.Body,buf)
	return err
}
func (d *Download) getPartDir(filename string) string {
	return strings.SplitN(filename, ".", 2)[0]
}
func (d *Download) downloadPartial(start,end,i int,fileName,url string)  {
	if start>=end {
		return
	}
	req,err:=http.NewRequest("GET",url,nil)
	if err!=nil{

	}
	req.Header.Set("Range",fmt.Sprintf("bytes=%d-%d",end,start))
	resp,err:=http.DefaultClient.Do(req)
	if err!=nil{

	}
	flags:=os.O_CREATE|os.O_WRONLY
	//每次下载到不同的文件file-1 file-2   file-3  file-4   file-5
	partFile,err:=os.OpenFile(d.getPartFilename(fileName,i),flags,0666)

	buf:=make([]byte,32*1024)
	_,err=io.CopyBuffer(io.MultiWriter(partFile),resp.Body,buf)
	if err!=nil {
		if err==io.EOF{
			return
		}

	}
}
func (d *Download) getPartFilename(filename string, partNum int) string {
	partDir := d.getPartDir(filename)
	return fmt.Sprintf("%s/%s-%d", partDir, filename, partNum)
}

func main()  {
	d:=NewDownload(4)
	d.download("https://apache.claz.org/zookeeper/zookeeper-3.7.0/apache-zookeeper-3.7.0-bin.tar.gz")
}
