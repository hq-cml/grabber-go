package downloader

import (
	"github.com/hq-cml/spider-go/middleware/pool"
	"reflect"
)

/********************************下载器池********************************/
//New,创建网页下载器
func NewDownloaderPool(total int, gen GenDownloaderFunc) (pool.PoolIntfs, error) {
	//直接调用gen()，利用反射获取期类型
	etype := reflect.TypeOf(gen())

	////gen()的返回值是DownloaderIntfs，但是它包含了pool.EntityIntfs所有声明方法
	////所以可以认为DownloaderIntfs是pool.EntityIntfs
	//genEntity := func() pool.EntityIntfs {
	//	return gen()
	//}
	pool, err := pool.NewPool(total, etype, gen)
	if err != nil {
		return nil, err
	}

	dl := &DownloaderPool{
		pool:  pool,
		etype: etype,
	}

	return dl, nil
}

//*DownloaderPool实现DownloaderPoolIntfs
func (dlpool *DownloaderPool) Get() (pool.EntityIntfs, error) {
	entity, err := dlpool.pool.Get()
	if err != nil {
		return nil, err
	}
	//dl, ok := entity.(DownloaderIntfs)
	//if !ok {
	//	msg := fmt.Sprintf("The type of entity is not %s!\n", dlpool.etype)
	//	panic(errors.New(msg))
	//}
    //
	//return dl, nil

	return entity, nil
}

func (dlpool *DownloaderPool) Put(dl pool.EntityIntfs) error {
	return dlpool.pool.Put(dl)
}

func (dlpool *DownloaderPool) Total() int {
	return dlpool.pool.Total()
}

func (dlpool *DownloaderPool) Used() int {
	return dlpool.pool.Used()
}

func (dlpool *DownloaderPool) Close() {
	dlpool.pool.Close()
}