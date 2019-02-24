package plugin

import (
	"testing"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"github.com/hq-cml/spider-go/helper/log"
)

func TestParseATag(t *testing.T) {
	log.InitLog("", "debug")

	resp, err := http.DefaultClient.Get("https://www.360.cn/")       //360首页，UTF8编码，content-type: text/html，没有指明charset
	//resp, err := http.DefaultClient.Get("http://www.dygang.net/")    //电影港首页，gbk编码，content-type: text/html，没有指明charset
	//resp, err := http.DefaultClient.Get("https://www.jianshu.com") //简书首页，UTF8编码，content-type: text/html; charset=utf-8
	if err != nil {
		t.Fatal(err)
	}

	items, _, errors := parseForATag(resp, 0, nil)

	//t.Log("分析出的URL列表:")
	//for _, req := range reqs {
	//	t.Logf("Depth: %d, URL: %s", req.Depth(), req.HttpReq().URL.String())
	//}

	t.Log("分析出的Item列表:", len(items))
	for _, item := range items {
		t.Log((*item)["url"])
		t.Log((*item)["charset"])
		t.Log((*item)["body"])
	}

	t.Log("分析出的Error列表:", len(errors))
	for _, err := range errors {
		t.Log(err)
	}
}

func TestGoQuery(t *testing.T) {
	html := `<body>
				<div lang="zh">DIV1</div>
				<p>P1</p>
				<div lang="zh-cn">DIV2</div>
				<div lang="en">DIV3</div>
				<span>
					<div style="display:none;">DIV4</div>
					<div>DIV5</div>
				</span>
				<p>P2</p>
				<div></div>
			</body>
			`

	dom,err:=goquery.NewDocumentFromReader(strings.NewReader(html))
	if err!=nil{
		t.Fatal(err)
	}

	dom.Find("body").Each(func(i int, selection *goquery.Selection) {
		t.Log(i, selection.Text())
	})
}

func TestHttpRequest(t *testing.T) {
	firstHttpReq, err := http.NewRequest("GET", "https://www.360.cn/", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(firstHttpReq.URL.Scheme)
}