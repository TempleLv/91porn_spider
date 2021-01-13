package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type VideoInfo struct {
	Title   string
	ViewKey string
	upTime  time.Time
	dlAddr  string
}

func (v VideoInfo) String() string {
	return fmt.Sprintf("VideoInfo: %s %s %s %s", v.Title, v.ViewKey, v.upTime.Format("2006-01-02 15:04:05"), v.dlAddr)
}

func (v *VideoInfo) updateDlAddr() (err error) {

	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", false),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		//chromedp.Flag("headless", false),
		//chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3830.0 Safari/537.36"),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, _ = context.WithTimeout(ctx, 60*time.Second)

	htmlText := ""
	fullUrl := "https://www.91porn.com/view_video.php?viewkey=" + v.ViewKey
	if err = chromedp.Run(ctx, sourHtml(fullUrl, "#player_one_html5_api > source", &htmlText)); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(htmlText)
	regAddr := regexp.MustCompile(`<source src="(?s:(.*?))" type="`)
	dlAddr := regAddr.FindAllStringSubmatch(htmlText, 1)
	if len(dlAddr) > 0 {
		v.dlAddr = dlAddr[0][1]
	}

	return
}

func (v VideoInfo) Download(savePath string, numThread int) (err error) {

	return err
}

func sourHtml(urlstr, sel string, html *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		//chromedp.WaitVisible(sel),
		//chromedp.Text("source src", html),
		//chromedp.ActionFunc(func(ctx context.Context) error {
		//	return nil
		//}),
		chromedp.OuterHTML(sel, html),
	}
}

func orgPageSave(dstUrl, proxyUrl, fileName string) {
	req, err := http.NewRequest("GET", dstUrl, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")

	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyUrl)
	}
	transport := &http.Transport{Proxy: proxy}
	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	buf, _ := ioutil.ReadAll(resp.Body)
	file.Write(buf)
}

func pageCrawl(dstUrl, proxyUrl string) (viAll []*VideoInfo) {
	req, err := http.NewRequest("GET", dstUrl, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")

	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyUrl)
	}
	transport := &http.Transport{Proxy: proxy}
	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	doc.Find("#wrapper > div.container.container-minheight > div.row > div > div > div > div").Each(func(i int, selection *goquery.Selection) {
		//fmt.Println(selection.Attr("href"))
		//fmt.Println(selection.HasClass(""))
		//fmt.Println(selection.Find("a").Attr("href"))
		//fmt.Println(selection.Find("a").Find("span.video-title").Text())
		textStr := selection.Text()

		selection.Find("a").Attr("href")
		title := selection.Find("a").Find("span.video-title").Text()
		videoUrl, urlOk := selection.Find("a").Attr("href")

		regViewKey := regexp.MustCompile(`91porn.com/view_video.php\?viewkey=(?s:(.*?))&page`)
		regAddTime := regexp.MustCompile(`添加时间:(?s:(.*?))\n`)

		viewkey := regViewKey.FindAllStringSubmatch(videoUrl, 1)
		addTime := regAddTime.FindAllStringSubmatch(textStr, 1)

		if len(viewkey) > 0 && len(addTime) > 0 && urlOk {

			vi := new(VideoInfo)

			title = ""
			strs := strings.Fields(addTime[0][1])

			if len(strs) == 3 {
				switch strs[1] {
				case "分钟":
					duration, _ := time.ParseDuration("-" + strs[0] + "m")
					vi.upTime = time.Now().Add(duration)
				case "小时":
					duration, _ := time.ParseDuration("-" + strs[0] + "h")
					vi.upTime = time.Now().Add(duration)
				case "天":
					hourDay, _ := strconv.Atoi(strs[0])
					hourDay = hourDay * 24
					duration, _ := time.ParseDuration("-" + strconv.Itoa(hourDay) + "h")
					vi.upTime = time.Now().Add(duration)
				}
			}
			vi.Title = title
			vi.ViewKey = viewkey[0][1]
			viAll = append(viAll, vi)
			//fmt.Println(vi)
		}

	})
	return
}

func main() {

	//testUrl := "https://www.91porn.com/view_video.php?viewkey=16b541c59e6efb52e8dd"
	const proxyUrl = "http://127.0.0.1:10808"
	//testUrl = "https://www.google.com/"

	viAll := pageCrawl("http://91porn.com/index.php", "http://127.0.0.1:10808")

	//orgPageSave(testUrl, proxyUrl, "1.html")
	for _, vi := range viAll {
		vi.updateDlAddr()
		fmt.Println(vi)
	}

	return
}
