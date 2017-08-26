package basic

/*
 * 基本数据类型的定义
 * *Request, *Response, Entry都是DataIntfs的实现
 */
import (
	"net/http"
)

/************************************** 数据类型接口 ************************************/
type DataIntfs interface {
	Valid() bool //数据是否有效
}

/************************************** Request 相关 **********************************/
//请求体结构
type Request struct {
	httpReq *http.Request //HTTP请求的指针，为了避免零值填充和实例复制，成员用指针
	depth   uint32        //请求深度，初始请求深度是0，然后逐渐递增
}

/**************************************** 响应体相关 ***********************************/
//响应体结构
type Response struct {
	httpResp *http.Response //HTTP响应的指针
	depth    uint32         //深度
}

/*************************************** 条目相关 **************************************/
//条目：一条响应经过分析之后的结果，因为处理链是定制的
//所以这个结构会尽量灵活以保证能够存储任意的分析结果
type Entry map[string]interface{}

/************************************** 错误类型相关 ************************************/
//错误类型
type ErrorType string

//错误类型常量
const (
	DOWNLOADER_ERROR      ErrorType = "Downloader Error"
	ANALYZER_ERROR        ErrorType = "Analyzer Error"
	ENTRY_PROCESSOR_ERROR ErrorType = "Entry Processor Error"
)

//错误类型
type SpiderError struct {
	errType    ErrorType //错误类型
	errMsg     string    //错误信息
	fullErrMsg string    //完整错误信息
}

/************************************** 参数类型相关 **********************************/
//Pool基本参数的容器。
type PoolParams struct {
	downloaderPoolSize uint32 // 网页下载器池的尺寸。
	analyzerPoolSize   uint32 // 分析器池的尺寸。
	description        string // 描述。
}

/************************************** 通道类型相关 **********************************/
type SpiderChannelIntfs interface{
	Put(data interface{}) error
	Get()(interface{}, bool)
	Len() int
	Cap() int
}

type RequestChannel struct {
	capacity      int
	reqCh         chan Request   //请求通道
}

type ResponseChannel struct {
	capacity      int
	respCh         chan Response   //响应通道
}

type EntryChannel struct {
	capacity      int
	entryCh         chan Entry   //结果通道
}

type ErrorChannel struct {
	capacity      int
	errorCh         chan SpiderError   //错误通道
}