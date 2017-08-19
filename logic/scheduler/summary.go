package scheduler

import (
    "fmt"
    "bytes"
    "strconv"
)

// 创建调度器摘要信息。
func NewSchedSummary(schdl *Scheduler, prefix string) SchedSummaryIntfs {
    if schdl == nil {
        return nil
    }
    urlCount := len(schdl.urlMap)
    var urlDetail string
    if urlCount > 0 {
        var buffer bytes.Buffer
        buffer.WriteByte('\n')
        for k, _ := range schdl.urlMap {
            buffer.WriteString(prefix)
            buffer.WriteString(prefix)
            buffer.WriteString(k)
            buffer.WriteByte('\n')
        }
        urlDetail = buffer.String()
    } else {
        urlDetail = "\n"
    }
    return &SchedSummary{
        prefix:              prefix,
        running:             schdl.running,
        //channelArgs:         sched.channelArgs,
        //poolBaseArgs:        sched.poolBaseArgs,
        poolSize:schdl.poolSize,
        channelLen:schdl.channelLen,
        grabDepth:          schdl.grabDepth,
        chanmanSummary:      schdl.channelManager.Summary(),
        reqCacheSummary:     schdl.requestCache.Summary(),
        dlPoolLen:           schdl.downloaderPool.Used(),
        dlPoolCap:           schdl.downloaderPool.Total(),
        analyzerPoolLen:     schdl.analyzerPool.Used(),
        analyzerPoolCap:     schdl.analyzerPool.Total(),
        processChainSummary: schdl.processChain.Summary(),
        urlCount:            urlCount,
        urlDetail:           urlDetail,
        stopSignSummary:     schdl.stopSign.Summary(),
    }
}


func (ss *SchedSummary) String() string {
    return ss.getSummary(false)
}

func (ss *SchedSummary) Detail() string {
    return ss.getSummary(true)
}

func (ss *SchedSummary) Same(other SchedSummary) bool {
    if other == nil {
        return false
    }
    otherSs, ok := interface{}(other).(*SchedSummary)
    if !ok {
        return false
    }
    if ss.running != otherSs.grabDepth ||
    ss.grabDepth != otherSs.grabDepth ||
    ss.dlPoolLen != otherSs.dlPoolLen ||
    ss.dlPoolCap != otherSs.dlPoolCap ||
    ss.analyzerPoolLen != otherSs.analyzerPoolLen ||
    ss.analyzerPoolCap != otherSs.analyzerPoolCap ||
    ss.urlCount != otherSs.urlCount ||
    ss.stopSignSummary != otherSs.stopSignSummary ||
    ss.reqCacheSummary != otherSs.reqCacheSummary ||
    //ss.poolBaseArgs.String() != otherSs.poolBaseArgs.String() ||
    //ss.channelArgs.String() != otherSs.channelArgs.String() ||
    ss.processChainSummary != otherSs.processChainSummary ||
    ss.chanmanSummary != otherSs.chanmanSummary {
        return false
    } else {
        return true
    }
}

// 获取摘要信息。
func (ss *SchedSummary) getSummary(detail bool) string {
    prefix := ss.prefix
    template := prefix + "Running: %v \n" +
    prefix + "Channel args: %s \n" +
    prefix + "Pool base args: %s \n" +
    prefix + "Crawl depth: %d \n" +
    prefix + "Channels manager: %s \n" +
    prefix + "Request cache: %s\n" +
    prefix + "Downloader pool: %d/%d\n" +
    prefix + "Analyzer pool: %d/%d\n" +
    prefix + "Item pipeline: %s\n" +
    prefix + "Urls(%d): %s" +
    prefix + "Stop sign: %s\n"
    return fmt.Sprintf(template,
        func() bool {
            return ss.running == 1
        }(),
        //ss.channelArgs.String(),
        //ss.poolBaseArgs.String(),
        strconv.Itoa(ss.channelLen),
        strconv.Itoa(ss.poolSize),
        ss.grabDepth,
        ss.chanmanSummary,
        ss.reqCacheSummary,
        ss.dlPoolLen, ss.dlPoolCap,
        ss.analyzerPoolLen, ss.analyzerPoolCap,
        ss.processChainSummary,
        ss.urlCount,
        func() string {
            if detail {
                return ss.urlDetail
            } else {
                return "<concealed>\n"
            }
        }(),
        ss.stopSignSummary)
}

