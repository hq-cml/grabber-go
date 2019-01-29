package scheduler

import (
    "fmt"
    "errors"
    "github.com/hq-cml/spider-go/basic"
    "github.com/hq-cml/spider-go/helper/log"
    "github.com/hq-cml/spider-go/middleware/pool"
    "github.com/hq-cml/spider-go/logic/analyzer"
    "github.com/hq-cml/spider-go/helper/util"
    "strings"
)

/*
 * 激活分析器，开始分析，分析工作由独立的goroutine进行负责，无限循环，从响应通道中获取响应
 * 每一个响应再交给独立的goroutine完成分析工作，分析goroutine整体上受到分析器池的制约
 */
func (schdl *Scheduler) activateAnalyzers(respAnalyzers []basic.AnalyzeResponseFunc) {
    go func() {
        for { //无限循环
            response, ok := schdl.getResponseChan().Get()
            if !ok {
                //通道已关闭
                break
            }
            resp, ok := response.(basic.Response)
            if !ok {
                continue
            }
            //启动异步分析
            go schdl.analyze(respAnalyzers, resp)
        }
    }()
}

//实际分析工作
func (schdl *Scheduler) analyze(respAnalyzers []basic.AnalyzeResponseFunc, response basic.Response) {
    //异常兜底
    defer func() {
        if p := recover(); p != nil {
            msg := fmt.Sprintf("Fatal Analysis Error: %s\n", p)
            log.Warn(msg)
        }
    }()

    entity, err := schdl.getAnalyzerPool().Get()
    if err != nil {
        msg := fmt.Sprintf("Analyzer pool error: %s", err)
        schdl.sendError(errors.New(msg), SCHEDULER_CODE)
        return
    }
    defer func() { //注册延时归还
        err = schdl.getAnalyzerPool().Put(entity)
        if err != nil {
            msg := fmt.Sprintf("Analyzer pool error: %s", err)
            schdl.sendError(errors.New(msg), SCHEDULER_CODE)
        }
    }()

    //断言转换
    ana, ok := entity.(*analyzer.Analyzer)
    if !ok {
        msg := fmt.Sprintf("Downloader pool Wrong type")
        schdl.sendError(errors.New(msg), SCHEDULER_CODE)
        return
    }
    moudleCode := generateModuleCode(ANALYZER_CODE, ana.Id())
    entryList, requestList, errs := ana.Analyze(respAnalyzers, response)

    //将分析出的entry放到entry通道里
    if entryList != nil {
        for _, entry := range entryList {
            schdl.sendEntry(*entry, moudleCode)
        }
    }

    //将分析出的request放到request缓冲
    if requestList != nil {
        for _, req := range requestList {
            if b := schdl.sendRequestToCache(*req, moudleCode); b {
                //标记URL已经扫过
                schdl.urlMap[req.HttpReq().URL.String()] = true
            }
        }
    }

    //将错误放到错误通道里
    if errs != nil {
        for _, err := range errs {
            schdl.sendError(err, moudleCode)
        }
    }
}

//发送条目到通道管理器中的条目通道
func (schdl *Scheduler) sendEntry(entry basic.Entry, moduleCode string) bool {
    if schdl.stopSign.Signed() {
        schdl.stopSign.Deal(moduleCode)
        return false
    }
    schdl.getEntryChan().Put(entry)
    return true
}

//获取Pool管理器持有的分析器Pool。
func (schdl *Scheduler) getAnalyzerPool() pool.PoolIntfs {
    p, err := schdl.poolManager.GetPool("analyzer")
    if err != nil {
        panic(err)
    }
    return p
}

//把请求存放到请求缓存。
func (schdl *Scheduler) sendRequestToCache(request basic.Request, mouduleCode string) bool {

    //过滤掉非法的请求
    if schdl.checkRequest(&request) == false {
        return false
    }

    //check停止信号
    if schdl.stopSign.Signed() {
        schdl.stopSign.Deal(mouduleCode)
        return false
    }

    schdl.requestCache.Put(&request)
    return true
}

//对分析出来的请求做合法性校验
func (schdl *Scheduler) checkRequest(request *basic.Request) bool {
    httpRequest := request.HttpReq()
    //校验请求体本身
    if httpRequest == nil {
        log.Warnln("Ignore the request! It's HTTP request is invalid!")
        return false
    }
    requestUrl := httpRequest.URL
    if requestUrl == nil {
        log.Warnln("Ignore the request! It's url is is invalid!")
        return false
    }

    if strings.ToLower(requestUrl.Scheme) != "http" {
        log.Warnf("Ignore the request! It's url is repeated. (requestUrl=%s)\n", requestUrl)
        return false
    }

    //已经处理过的URL不再处理
    if _, ok := schdl.urlMap[requestUrl.String()]; ok {
        log.Warnf("Ignore the request! It's url is repeated. (requestUrl=%s)\n", requestUrl)
        return false
    }

    //只有主域名相同的URL才是合法的
    //TODO 这个地方可以不一定
    if pd, _ := util.GetPrimaryDomain(httpRequest.Host); pd != schdl.primaryDomain {
        log.Warnf("Ignore the request! It's host '%s' not in primary domain '%s'. (requestUrl=%s)\n",
            httpRequest.Host, schdl.primaryDomain, requestUrl)
        return false
    }

    //请求深度不能超过阈值
    if request.Depth() > schdl.grabMaxDepth {
        log.Warnf("Ignore the request! It's depth %d greater than %d. (requestUrl=%s)\n",
            request.Depth(), schdl.grabMaxDepth, requestUrl)
        return false
    }
    return true
}
