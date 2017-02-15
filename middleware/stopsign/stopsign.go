package stopsign

import (
    "sync"
    "fmt"
)

/*
 * 停止信号
 * 实现"一方发送，多方接收的效果"
 */

//停止信号接口类型
type StopSignIntfs interface {
    Sign() bool //发出停止信号。如果先前已发出过停止信号，那么该方法会返回false。
    Signed() bool// 判断停止信号是否已被发出。
    Reset()// 重置停止信号。相当于收回停止信号，并清除所有的停止信号处理记录。
    Deal(code string)// 处理停止信号。当处理了停止信号之后，停止信号的处理方应该调用停止信号Deal方法，表示已经对该信号处理完毕。参数code应该代表停止信号处理方的代号。
    DealCount(code string) uint32// 获取某一个停止信号处理方的处理计数。
    DealTotal() uint32// 获取停止信号被处理的总计数。
    Summary() string// 获取摘要信息。其中应该包含所有的停止信号处理记录。
}

//停止信号实现
type StopSign struct {
    rwmutex   sync.RWMutex         //保护锁
    signed    bool                 //信号是否已发送标志
    dealCountMap map[string]uint32 //处理方处理计数map
}

//惯例New
func NewStopSign() StopSignIntfs {
    return &StopSign{
        dealCountMap: make(map[string]uint32),
    }
}

//*StopSign实现StopSignIntfs接口
func (s *StopSign) Sign() bool {
    s.rwmutex.Lock()
    defer s.rwmutex.Unlock()
    if s.signed {
        return false
    }
    s.signed = true
    return true
}

func (s *StopSign) Signed() bool {
    return s.signed
}

func (s *StopSign) Reset() {
    s.rwmutex.Lock()
    defer s.rwmutex.Unlock()
    s.signed = false
    //直接重置了，扔给GC去回收了
    s.dealCountMap = make(map[string]uint32)
}

func (s *StopSign) Deal (code string) {
    s.rwmutex.Lock()
    defer s.rwmutex.Unlock()

    if !s.signed {
        return
    }

    if _, ok := s.dealCountMap[code]; !ok {
        s.dealCountMap[code] = 1
    } else {
        s.dealCountMap[code] += 1
    }
}

func (s *StopSign) DealCount(code string) uint32 {
    s.rwmutex.RLock()
    defer s.rwmutex.RUnlock()
    return s.dealCountMap[code]
}

func (s *StopSign) DealTotal() uint32 {
    s.rwmutex.RLock()
    defer s.rwmutex.RUnlock()
    var total uint32
    for _,v := range s.dealCountMap {
        total += v
    }
    return total
}

func (s *StopSign) Summary() string {
    if s.signed {
        return fmt.Sprintf("signed: true, dealCount: %v", s.dealCountMap)
    } else {
        return "Signed: false"
    }
}


















