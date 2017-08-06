package downloader

import (
    "net/http"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/middleware/pool"
    "reflect"
)

/*
 * 下载器主要职责是下载网页
 */
type DownloaderIntfs interface {
    Id() uint64 //获得下载器的Id
    Download(req basic.Request) (*basic.Response, error) //实际的下载行为
}

//网页下载器，*Downloader实现PageDownloaderIntfs接口
type Downloader struct {
    id  uint64 //ID
    httpClient http.Client
}

/*
 * 网页下载器存在于下载器池中，每个下载器有自己的Id
 */
//生成网页下载器的函数的类型
type GenDownloader func() DownloaderIntfs

//下载器接口类型
type DownloaderPoolIntfs interface {
    Get() (DownloaderIntfs, error)      // 从池中获取一个下载器
    Put(dl DownloaderIntfs) error       // 归还一个下载器到池子中
    Total() uint32                      // 获得池子总容量
    Used() uint32                       // 获得正在被使用的下载器数量
}

//网页下载器池的实现，*DownloaderPool实现DownloaderPoolIntfs
type DownloaderPool struct {
    pool  pool.Pool //结构体的嵌套
    etype reflect.Type
}
