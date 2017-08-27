package scheduler

/*
 * 监视器实现：主要功能是对Scheduler的监视和控制：
 * 1. 在适当的时候停止自身和Scheduler
 * 2. 实时监控Scheduler及其各个组件的运行状况
 * 3. 一旦Scheduler及其各组件发生错误能够及时报告
 */
import (
	"errors"
	"fmt"
	"runtime"
	"time"
)

// 日志记录函数的类型。
// 参数level代表日志级别。级别设定：0：普通；1：警告；2：错误。
type Record func(level byte, content string)

// 调度器监控函数。
// 参数scheduler代表作为监控目标的调度器。
// 参数intervalNs代表检查间隔时间，单位：纳秒。
// 参数maxIdleCount代表最大空闲计数。
// 参数autoStop被用来指示该方法是否在调度器空闲一段时间（即持续空闲时间，由intervalNs * maxIdleCount得出）之后自行停止调度器。
// 参数detailSummary被用来表示是否需要详细的摘要信息。
// 参数record代表日志记录函数。
// 当监控结束之后，该方法会会向作为唯一返回值的通道发送一个代表了空闲状态检查次数的数值。
func Monitoring(
	scheduler SchedulerIntfs,
	intervalNs time.Duration,
	maxIdleCount uint,
	autoStop bool,
	detailSummary bool,
	record Record) <-chan uint64 {

	if scheduler == nil { // 调度器不能不可用！
		panic(errors.New("The scheduler is invalid!"))
	}
	// 防止过小的参数值对爬取流程的影响
	if intervalNs < time.Millisecond {
		intervalNs = time.Millisecond
	}
	if maxIdleCount < 1000 {
		maxIdleCount = 1000
	}

	// 监控停止通知器
	stopNotifier := make(chan byte, 1)

	//从错误通道中接收和报告错误。
	reportError(scheduler, record, stopNotifier)

	//记录摘要信息
	recordSummary(scheduler, detailSummary, record, stopNotifier)

	//检查计数通道
	checkCountChan := make(chan uint64, 2)

	//检查空闲状态
	checkStatus(scheduler, intervalNs, maxIdleCount, autoStop, checkCountChan, record, stopNotifier)
	return checkCountChan
}

// 检查状态，并在满足持续空闲时间的条件时采取必要措施。
func checkStatus(
	scheduler SchedulerIntfs,
	intervalNs time.Duration,
	maxIdleCount uint,
	autoStop bool,
	checkCountChan chan<- uint64,
	record Record,
	stopNotifier chan<- byte) {

	var checkCount uint64
	// 已达到最大空闲计数的消息模板。
	var msgReachMaxIdleCount = "The scheduler has been idle for a period of time (about %s). \n" +
		"Now consider what stop it."
	// 停止调度器的消息模板。
	var msgStopScheduler = "Stop scheduler...%s."

	go func() {
		defer func() {
			stopNotifier <- 1
			stopNotifier <- 2
			checkCountChan <- checkCount
		}()
		// 等待调度器开启
		waitForSchedulerStart(scheduler)
		// 准备
		var idleCount uint
		var firstIdleTime time.Time
		for {
			// 检查调度器的空闲状态
			if scheduler.Idle() {
				idleCount++
				if idleCount == 1 {
					firstIdleTime = time.Now()
				}
				if idleCount >= maxIdleCount {
					msg := fmt.Sprintf(msgReachMaxIdleCount, time.Since(firstIdleTime).String())
					record(0, msg)
					// 再次检查调度器的空闲状态，确保它已经可以被停止
					if scheduler.Idle() {
						if autoStop {
							var result string
							if scheduler.Stop() {
								result = "success"
							} else {
								result = "failing"
							}
							msg = fmt.Sprintf(msgStopScheduler, result)
							record(0, msg)
						}
						break
					} else {
						if idleCount > 0 {
							idleCount = 0
						}
					}
				}
			} else {
				if idleCount > 0 {
					idleCount = 0
				}
			}
			checkCount++
			time.Sleep(intervalNs)
		}
	}()
}

// 记录摘要信息。
func recordSummary(
	scheduler SchedulerIntfs,
	detailSummary bool,
	record Record,
	stopNotifier <-chan byte) {

	// 摘要信息的模板。
	var summaryForMonitoring = "Monitor - Collected information[%d]:\n" +
		"  Goroutine number: %d\n" +
		"  Scheduler:\n%s" +
		"  Escaped time: %s\n"

	go func() {
		//阻塞等待调度器开启
		waitForSchedulerStart(scheduler)

		// 准备
		var prevSchedSummary SchedSummaryIntfs
		var prevNumGoroutine int
		var recordCount uint64 = 1
		startTime := time.Now()

		for {
			// 查看监控停止通知器
			select {
			case <-stopNotifier:
				return
			default:
			}
			// 获取摘要信息的各组成部分
			currNumGoroutine := runtime.NumGoroutine()
			currSchedSummary := scheduler.Summary("    ")
			// 比对前后两份摘要信息的一致性。只有不一致时才会予以记录。主要为了防止日志的大量生产造成干扰
			if currNumGoroutine != prevNumGoroutine || !currSchedSummary.Same(prevSchedSummary) {
				schedSummaryStr := func() string {
					if detailSummary {
						return currSchedSummary.Detail()
					} else {
						return currSchedSummary.String()
					}
				}()
				// 记录摘要信息
				info := fmt.Sprintf(summaryForMonitoring,
					recordCount,
					currNumGoroutine,
					schedSummaryStr,
					time.Since(startTime).String(), //当前时间和startTime的时间间隔
				)
				record(0, info)
				prevNumGoroutine = currNumGoroutine
				prevSchedSummary = currSchedSummary
				recordCount++
			}
			//time.Sleep(time.Microsecond)
			time.Sleep(time.Second)
		}
	}()
}

//从错误通道中接收和报告错误。
func reportError(
	scheduler SchedulerIntfs,
	record Record,
	stopNotifier <-chan byte) {

	go func() {
		//阻塞等待调度器开启
		waitForSchedulerStart(scheduler)
		for {
			//非阻塞得查看监控停止通知器
			select {
			case <-stopNotifier:
				return
			default: //非阻塞
			}

			err, ok := scheduler.ErrorChan().Get()
			if !ok {
				return
			}
			//如果errorChan关闭，则err可能是nil
			if err != nil {
				errMsg := fmt.Sprintf("Error (received from error channel): %s", err)
				record(2, errMsg)
			}
			//让出时间片
			time.Sleep(time.Microsecond)
		}
	}()
}

//阻塞等待调度器开启。
func waitForSchedulerStart(scheduler SchedulerIntfs) {
	for !scheduler.Running() {
		time.Sleep(time.Microsecond)
	}
}