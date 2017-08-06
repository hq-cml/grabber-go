package channelmanager

import (
    "github.com/hq-cml/spider-go/basic"
    "errors"
    "fmt"
)

//New
func NewChannelManager(channelLen uint) ChannelManagerIntfs {
    if channelLen == 0 {
        channelLen = 1024
    }
    chm := &ChannelManager{}
    chm.Init(channelLen, true)
    return chm
}

// 检查状态，保证在获取通道的时候，通道管理器应处于已初始化状态。
func (chm *ChannelManager) checkStatus() error {
    if chm.status == CHANNEL_MANAGER_STATUS_INITIALIZED {
        return nil
    }
    statusName, ok := statusNameMap[chm.status]
    if !ok {
        statusName = fmt.Sprintf("%d", chm.status)
    }
    errMsg := fmt.Sprintf("the undesirable status of channel manager :%s!\n", statusName)
    return errors.New(errMsg)
}

//*ChannelManager实现ChannelManagerIntfs接口
//Init方法
func (chm *ChannelManager) Init(channelLen uint, reset bool) bool {
    if channelLen == 0 {
        panic(errors.New("The channel length is invalid!"))
    }
    //写锁保护
    chm.rwmutex.Lock()
    defer chm.rwmutex.Unlock()

    //避免重复初始化
    if chm.status == CHANNEL_MANAGER_STATUS_INITIALIZED && reset != true {
        return false
    }
    chm.channelLen = channelLen
    chm.reqCh = make(chan basic.Request, channelLen)
    chm.respCh = make(chan basic.Response, channelLen)
    chm.itemCh = make(chan basic.Item, channelLen)
    chm.errorCh = make(chan error, channelLen)
    chm.status = CHANNEL_MANAGER_STATUS_INITIALIZED

    return true
}

//close关闭
func (chm *ChannelManager) Close() bool {
    //写锁保护
    chm.rwmutex.Lock()
    defer chm.rwmutex.Unlock()

    if chm.status != CHANNEL_MANAGER_STATUS_INITIALIZED {
        return false
    }

    close(chm.reqCh)
    close(chm.respCh)
    close(chm.itemCh)
    close(chm.errorCh)
    chm.status = CHANNEL_MANAGER_STATUS_CLOSED

    return true
}

//获取request通道
func (chanman *ChannelManager) ReqChan() (chan basic.Request, error) {
    //读锁保护
    chanman.rwmutex.RLock()
    defer chanman.rwmutex.RUnlock()
    if err := chanman.checkStatus(); err != nil {
        return nil, err
    }
    return chanman.reqCh, nil
}

//获取response通道
func (chanman *ChannelManager) RespChan() (chan basic.Response, error) {
    //读锁保护
    chanman.rwmutex.RLock()
    defer chanman.rwmutex.RUnlock()
    if err := chanman.checkStatus(); err != nil {
        return nil, err
    }
    return chanman.respCh, nil
}

//获取item通道
func (chanman *ChannelManager) ItemChan() (chan basic.Item, error) {
    //读锁保护
    chanman.rwmutex.RLock()
    defer chanman.rwmutex.RUnlock()
    if err := chanman.checkStatus(); err != nil {
        return nil, err
    }
    return chanman.itemCh, nil
}

//获取error通道
func (chanman *ChannelManager) ErrorChan() (chan error, error) {
    //读锁保护
    chanman.rwmutex.RLock()
    defer chanman.rwmutex.RUnlock()
    if err := chanman.checkStatus(); err != nil {
        return nil, err
    }
    return chanman.errorCh, nil
}

//摘要方法
func (chanman *ChannelManager) Summary() string {
    //模板
    chanmanSummaryTemplate := "status: %s, " +
    "requestChannel: %d/%d, " +
    "responseChannel: %d/%d, " +
    "itemChannel: %d/%d, " +
    "errorChannel: %d/%d"

    summary := fmt.Sprintf(chanmanSummaryTemplate,
        statusNameMap[chanman.status],
        len(chanman.reqCh), cap(chanman.reqCh),
        len(chanman.respCh), cap(chanman.respCh),
        len(chanman.itemCh), cap(chanman.itemCh),
        len(chanman.errorCh), cap(chanman.errorCh))
    return summary
}












