package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

func (v *VideoInfo) updateDlAddr(proxy string) (err error) {
	v.dlAddr = ""
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", false),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.ProxyServer(proxy),
		//chromedp.Flag("headless", false),
		//chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3830.0 Safari/537.36"),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, _ = context.WithTimeout(ctx, time.Minute)

	htmlText := ""
	fullUrl := "https://www.91porn.com/view_video.php?viewkey=" + v.ViewKey
	if err = chromedp.Run(ctx, sourHtml(fullUrl, "#player_one_html5_api > source", &htmlText)); err != nil {
		fmt.Println(err)
		return
	}
	regAddr := regexp.MustCompile(`<source src="(?s:(.*?))" type="`)
	dlAddr := regAddr.FindAllStringSubmatch(htmlText, 1)
	if len(dlAddr) > 0 {
		v.dlAddr = dlAddr[0][1]
	}

	return
}

func (v VideoInfo) Download(savePath string, numThread int, proxy string) (err error) {

	if len(v.dlAddr) > 0 {
		//strCmd := fmt.Sprintf(" -p \"%s\" -t %d -w -o %s \"%s\"", proxy, numThread, savePath, v.dlAddr)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "m3_dl", "-p", proxy, "-t", strconv.Itoa(numThread), "-w", "-o", savePath, v.dlAddr)

		//cmd := exec.Command("m3_dl", "-p", proxy, "-t", strconv.Itoa(numThread), "-w", "-o", savePath, v.dlAddr)
		//fmt.Println(cmd)
		out, ierr := cmd.CombinedOutput()
		if ierr != nil {
			//fmt.Println(string(out))
			_ = out
			err = ierr
			fmt.Println(v.Title, "download fail!")
		} else {
			fmt.Println(v.Title, "download success!")
		}
	} else {
		fmt.Println(v.Title, "dlAddr not set!")
		err = fmt.Errorf(v.Title, "dlAddr not set!")
	}

	return
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

func pageCrawl(dstUrl, proxyUrl string) (viAll []*VideoInfo) {
	req, err := http.NewRequest("GET", dstUrl, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")

	client := &http.Client{}

	if len(proxyUrl) > 0 {
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(proxyUrl)
		}
		transport := &http.Transport{Proxy: proxy}
		client = &http.Client{Transport: transport}
	}

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

			//title = "1"
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

func orgPageSave(dstUrl, proxyUrl, fileName string) {
	req, err := http.NewRequest("GET", dstUrl, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")

	client := &http.Client{}

	if len(proxyUrl) > 0 {
		fmt.Println("use proxy")
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(proxyUrl)
		}
		transport := &http.Transport{Proxy: proxy}
		client = &http.Client{Transport: transport}
	}

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

func DownladMany(viAll []*VideoInfo, numThread int, proxyUrl, savePath string) {
	ch := make(chan int, len(viAll))
	chq := make(chan int, numThread)
	fmt.Print("DownladMany:len([]*VideoInfo)=", len(viAll), "\n")
	for _, vi := range viAll {
		go func(info *VideoInfo) {
			chq <- 1
			info.updateDlAddr(proxyUrl)
			savePath := filepath.Join(savePath, fmt.Sprintf("%s.ts", info.Title))
			info.Download(savePath, 20, proxyUrl)
			<-chq
			ch <- 1
		}(vi)
		time.Sleep(time.Second * 2)
	}

	for range viAll {
		<-ch
	}
}

func main() {

	proxyUrl := ""
	pageUrl := ""
	savePath := ""
	threadNum := 5

	flag.StringVar(&proxyUrl, "p", "http://192.168.4.66:10808", "proxy")
	flag.StringVar(&pageUrl, "u", "http://91porn.com/index.php", "page to crawl")
	flag.StringVar(&savePath, "o", "./save", "path to output")
	flag.IntVar(&threadNum, "t", 5, "thradcount")

	flag.Parse()

	path, _ := filepath.Abs(savePath)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			fmt.Println(err)
		}
	}

	viAll := pageCrawl(pageUrl, proxyUrl)

	DownladMany(viAll, threadNum, proxyUrl, path)

	return
}
