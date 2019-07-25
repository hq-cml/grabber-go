package plugin

import (
	"github.com/hq-cml/spider-man/basic"
	"net/http"
	"time"
	"net/url"
	"io"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

/*
 * *EngineSpider实现SpiderPlugin接口
 * 此插件与搜索引擎Spider-Engine打通，爬虫爬取结果=>灌入Spider-Engine
 */
type EngineSpider struct {
	userData interface{}
}

//New
func NewEngineSpider(v interface{}) basic.SpiderPlugin {
	return &EngineSpider{
		userData: v,
	}
}

//*EngineSpider实现SpiderPlugin接口
//生成HTTP客户端
func (b *EngineSpider) GenHttpClient() *http.Client {
	//客户端必须设置一个整体超时时间，否则随着时间推移，会把downloader全部卡死
	return &http.Client{
		Transport: http.DefaultTransport,
		Timeout: time.Duration(basic.Conf.RequestTimeout) * time.Second,
	}
}

//获得响应解析函数的序列
func (b *EngineSpider) GenResponseAnalysers() []basic.AnalyzeResponseFunc {
	return []basic.AnalyzeResponseFunc {
		parse360NewsPage,
	}
}

// 获得条目处理链的序列。
func (b *EngineSpider) GenItemProcessors() []basic.ProcessItemFunc {
	return []basic.ProcessItemFunc{
		//闭包
		func(item basic.Item) (basic.Item, error) {
			addr, ok := b.userData.(string)
			if !ok {
				panic("Wrong type")
			}
			return processEngineItem(item, addr)
		},

	}
}

// 页面分析, 通过分析360新闻页面的Dom元素，爬取规则自然也就完成了
// 针对360的新闻页面进行分析
// 360新闻首页：http://www.360.cn/news.html
// 常规新闻URL：http://www.360.cn/n/10758.html
func parse360NewsPage(httpResp *basic.Response) ([]*basic.Item, []*basic.Request, []error) {

	//对响应做一些处理
	reqUrl, err := url.Parse(httpResp.ReqUrl) //记录下响应的请求（防止相对URL的问题）
	if err != nil {
		return nil, nil, []error{err}
	}
	var httpRespBody io.Reader

	itemList := []*basic.Item{}
	requestList := []*basic.Request{}
	errs := make([]error, 0)

	//网页编码智能判断, 非utf8 => utf8
	httpRespBody, contentType, err := convertCharset(httpResp)
	if err != nil {
		return nil, nil, []error{err}
	}

	//开始解析
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return nil, nil, errs
	}

	//查找“A”标签并提取链接地址
	requestList, errs = findATagFromDoc(httpResp, reqUrl, doc)

	//关键字查找, 记录符合条件的body作为item
	//如果用户数据非空，则进行匹配校验，否则直接入item队列
	imap := make(map[string]interface{})
	imap["url"] = reqUrl.String()
	imap["charset"] = contentType
	imap["depth"] = httpResp.Depth
	imap["title"] = strings.TrimRight(strings.TrimLeft(doc.Find(".article-content").Find("h1").Text(), " \n"), " \n")
	imap["time"] = strings.TrimRight(strings.TrimLeft(doc.Find(".article-content").Find(".article-info").Find("ul").Find("li").First().Text(), " \n"), " \n")
	imap["content"] = strings.TrimRight(strings.TrimLeft(doc.Find(".article-content").Find(".content-text").Text(), " \n"), " \n")
	item := basic.Item(imap)
	itemList = append(itemList, &item)

	return itemList, requestList, errs
}

// 条目处理函数
// 发送到Spider-Engine
func processEngineItem(item basic.Item, engineAddr string) (result basic.Item, err error) {


	return nil, nil
}