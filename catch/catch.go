package catch

import (
	"bytes"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
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
	"sync"
	"time"
)

type VideoInfo struct {
	Title   string
	ViewKey string
	Owner   string
	UpTime  time.Time
	DlAddr  string
	Vdurat  float64
	Watch   int
	Collect int
	Score   float64
}

type ViSlice []*VideoInfo

func (v ViSlice) Len() int { return len(v) }

func (v ViSlice) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

func (v ViSlice) Less(i, j int) bool { return v[i].Score < v[j].Score }

func (v ViSlice) String() string {
	str := ""
	for _, v := range v {
		str += fmt.Sprintf("VideoInfo: (%.1f %.1f)%s %s %d %d %s %s\n",
			v.Score, v.Vdurat, v.Title, v.ViewKey, v.Watch, v.Collect, v.UpTime.Format("2006-01-02 15:04:05"), v.DlAddr)
	}

	return str
}

func (v VideoInfo) String() string {
	return fmt.Sprintf("VideoInfo: (%.1f %.1f)%s %s %d %d %s %s",
		v.Score, v.Vdurat, v.Title, v.ViewKey, v.Watch, v.Collect, v.UpTime.Format("2006-01-02 15:04:05"), v.DlAddr)
}

func (v *VideoInfo) updateDlAddr(proxy string) (err error) {
	v.DlAddr = ""
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", false),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.ProxyServer(proxy),
		//chromedp.Flag("headless", false),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0"),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, _ = context.WithTimeout(ctx, time.Second*10)

	htmlText := ""
	fullUrl := "https://www.91porn.com/view_video.php?viewkey=" + v.ViewKey
	if err = chromedp.Run(ctx, sourHtml(fullUrl, "#player_one_html5_api > source", &htmlText)); err != nil {
		if err = chromedp.Run(ctx, sourHtml_click(fullUrl, "#player_one_html5_api > source", &htmlText)); err != nil {
			fmt.Println("DlAddr", fullUrl, err)
			return
		} else {
			fmt.Println(fullUrl, "DlAddr done!")
		}
	} else {
		fmt.Println(fullUrl, "DlAddr done!")
	}
	regAddr := regexp.MustCompile(`<source src="(?s:(.*?))" type="`)
	dlAddr := regAddr.FindAllStringSubmatch(htmlText, 1)
	if len(dlAddr) > 0 {
		v.DlAddr = dlAddr[0][1]
		v.DlAddr = strings.ReplaceAll(v.DlAddr, "&amp;", "&")
	}

	return
}

func (v VideoInfo) Download(savePath string, numThread int, proxy string) (err error) {

	if len(v.DlAddr) > 0 {
		//strCmd := fmt.Sprintf(" -p \"%s\" -t %d -w -o %s \"%s\"", proxy, numThread, savePath, v.DlAddr)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5+time.Second*30)
		defer cancel()
		cmd := exec.CommandContext(ctx, "m3_dl", "-p", proxy, "-t", strconv.Itoa(numThread), "-w", "-o", savePath, v.DlAddr)

		//cmd := exec.Command("m3_dl", "-p", proxy, "-t", strconv.Itoa(numThread), "-w", "-o", savePath, v.DlAddr)
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
		fmt.Println(v.Title, "DlAddr not set!")
		err = fmt.Errorf(v.Title, "DlAddr not set!")
	}

	return
}

func sourHtml(urlstr, sel string, html *string) chromedp.Tasks {
	return chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{
			"Accept-Language": "zh-CN,zh;q=0.9",
		})),
		chromedp.Navigate(urlstr),
		//chromedp.WaitVisible(sel),
		//chromedp.Text("source src", html),
		//chromedp.ActionFunc(func(ctx context.Context) error {
		//	return nil
		//}),
		//chromedp.Click("body > table > tbody > tr > td > a", chromedp.ByQuery),
		chromedp.OuterHTML(sel, html),
	}
}

func sourHtml_click(urlstr, sel string, html *string) chromedp.Tasks {
	return chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{
			"Accept-Language": "zh-CN,zh;q=0.9",
		})),
		chromedp.Navigate(urlstr),
		//chromedp.WaitVisible(sel),
		//chromedp.Text("source src", html),
		//chromedp.ActionFunc(func(ctx context.Context) error {
		//	return nil
		//}),
		chromedp.Click("body > table > tbody > tr > td > a", chromedp.ByQuery),
		chromedp.OuterHTML(sel, html),
	}
}

func sourManyHtml(urlstr string, sel, html []string) chromedp.Tasks {
	task := chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{
			"Accept-Language": "zh-CN,zh;q=0.9",
		})),
		chromedp.Navigate(urlstr),
	}
	if len(sel) == len(html) {
		for i, _ := range sel {
			task = append(task, chromedp.OuterHTML(sel[i], &html[i]))
		}
	}

	return task
}

func nopCrawHtml(urlstr string, sel string, html *string) chromedp.Tasks {
	task := chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{
			"Accept-Language": "zh-CN,zh;q=0.9",
		})),
		chromedp.Navigate(urlstr),
		//chromedp.Click("body > table > tbody > tr > td > a", chromedp.ByQuery),
		chromedp.OuterHTML(sel, html),
	}

	return task
}

func nopCrawHtml_click(urlstr string, sel string, html *string) chromedp.Tasks {
	task := chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{
			"Accept-Language": "zh-CN,zh;q=0.9",
		})),
		chromedp.Navigate(urlstr),
		chromedp.Click("body > table > tbody > tr > td > a", chromedp.ByQuery),
		chromedp.OuterHTML(sel, html),
	}

	return task
}

func PageCrawlOne(dstUrl, proxyUrl string) (vi VideoInfo, err error) {
	vi.DlAddr = ""
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"),
		chromedp.ProxyServer(proxyUrl),
		//chromedp.Flag("headless", false),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0"),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, _ = context.WithTimeout(ctx, time.Second*25)

	sels := [...]string{"#player_one_html5_api > source", "#videodetails > h4", "#videodetails-content > div:nth-child(2) > span.title-yakov > a:nth-child(1) > span"}
	htmlText := [len(sels)]string{}
	if err = chromedp.Run(ctx, sourManyHtml(dstUrl, sels[:], htmlText[:])); err != nil {
		fmt.Println(err)
		return
	}
	regAddr := regexp.MustCompile(`<source src="(?s:(.*?))" type="`)
	regTitle := regexp.MustCompile(`">(?s:(.*?))</h4>`)
	regOwner := regexp.MustCompile(`">(?s:(.*?))</span>`)
	dlAddr := regAddr.FindAllStringSubmatch(htmlText[0], 1)
	title := regTitle.FindAllStringSubmatch(htmlText[1], 1)
	owner := regOwner.FindAllStringSubmatch(htmlText[2], 1)
	if len(dlAddr) > 0 && len(title) > 0 && len(owner) > 0 {
		vi.DlAddr = dlAddr[0][1]
		//将vi.DlAddr中的&amp;转换为&
		vi.DlAddr = strings.ReplaceAll(vi.DlAddr, "&amp;", "&")
		//去掉title[0][1]中的空格和换行符
		vi.Title = strings.ReplaceAll(title[0][1], " ", "")
		vi.Title = strings.ReplaceAll(vi.Title, "\n", "")
		vi.Owner = owner[0][1]
	}

	return
}

func PageCrawl_chromedp(dstUrl, proxyUrl string) (viAll []*VideoInfo) {
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", false),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.ProxyServer(proxyUrl),
		//chromedp.Flag("headless", false),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, _ = context.WithTimeout(ctx, time.Second*25)

	sel := "#wrapper"
	htmlText := ""
	if err := chromedp.Run(ctx, nopCrawHtml(dstUrl, sel, &htmlText)); err != nil {
		if err := chromedp.Run(ctx, nopCrawHtml_click(dstUrl, sel, &htmlText)); err != nil {
			fmt.Println("Crawl", dstUrl, err)
			return
		} else {
			fmt.Println(dstUrl, "Crawl done!")
		}
	} else {
		fmt.Println(dstUrl, "Crawl done!")
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(htmlText))
	if err != nil {
		fmt.Println(err)
		return
	}

	doc.Find("#wrapper > div.container.container-minheight > div.row > div > div > div > div").Each(func(i int, selection *goquery.Selection) {
		textStr := selection.Text()

		title := selection.Find("a").Find("span.video-title").Text()
		videoUrl, urlOk := selection.Find("a").Attr("href")

		regViewKey := regexp.MustCompile(`viewkey=(?s:(.*))&page`)
		regAddTime := regexp.MustCompile(`添加时间:(?s:(.*?))\n`)
		regWatch := regexp.MustCompile(`热度:(?s:(.*?))\n`)
		regCollect := regexp.MustCompile(`收藏:(?s:(.*?))\n`)
		regOwner := regexp.MustCompile(`作者: \n(?s:(.*?))\n`)

		viewkey := regViewKey.FindAllStringSubmatch(videoUrl, 1)
		addTime := regAddTime.FindAllStringSubmatch(textStr, 1)
		watch := regWatch.FindAllStringSubmatch(textStr, 1)
		collect := regCollect.FindAllStringSubmatch(textStr, 1)
		owner := regOwner.FindAllStringSubmatch(textStr, 1)

		if len(viewkey) > 0 && len(addTime) > 0 && len(watch) > 0 && len(collect) > 0 && len(owner) > 0 && urlOk {

			vi := new(VideoInfo)

			//title = "1"
			strs := strings.Fields(addTime[0][1])

			if len(strs) == 3 {
				switch strs[1] {
				case "分钟":
					duration, _ := time.ParseDuration("-" + strs[0] + "m")
					vi.UpTime = time.Now().Add(duration)
				case "小时":
					duration, _ := time.ParseDuration("-" + strs[0] + "h")
					vi.UpTime = time.Now().Add(duration)
				case "天":
					hourDay, _ := strconv.Atoi(strs[0])
					hourDay = hourDay * 24
					duration, _ := time.ParseDuration("-" + strconv.Itoa(hourDay) + "h")
					vi.UpTime = time.Now().Add(duration)
				}
			}
			vi.Title = title
			vi.ViewKey = viewkey[0][1]
			strs = strings.Fields(watch[0][1])
			vi.Watch, _ = strconv.Atoi(strs[0])
			strs = strings.Fields(collect[0][1])
			if len(strs) > 0 {
				vi.Collect, _ = strconv.Atoi(strs[0])
			}

			if len(strings.Fields(owner[0][1])) > 0 {
				vi.Owner = strings.Fields(owner[0][1])[0]
			} else {
				vi.Owner = "unknown"
			}

			vMinute := 0
			vSecond := 0
			fmt.Sscanf(selection.Find("span.duration").Text(), "%d:%d\n", &vMinute, &vSecond)
			vi.Vdurat = float64(vMinute) + float64(vSecond)/60.0

			viAll = append(viAll, vi)

			//fmt.Println(vi)
		}

	})

	return
}

func PageCrawl(dstUrl, proxyUrl string) (viAll []*VideoInfo) {
	req, err := http.NewRequest("GET", dstUrl, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Add("sec-ch-ua", "\" Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("Sec-Fetch-Site", "none")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Accept-Encoding", " gzip, deflate, br")
	//req.Header.Add("", "")

	//req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")

	//req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	//req.Header.Add("Accept-Encoding", "gzip, deflate")
	//req.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")

	//req.Header.Add("Upgrade-Insecure-Requests", "1")
	//req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

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

		title := selection.Find("a").Find("span.video-title").Text()
		videoUrl, urlOk := selection.Find("a").Attr("href")

		regViewKey := regexp.MustCompile(`91porn.com/view_video.php\?viewkey=(?s:(.*?))&page`)
		regAddTime := regexp.MustCompile(`添加时间:(?s:(.*?))\n`)
		regWatch := regexp.MustCompile(`查看:(?s:(.*?))\n`)
		regCollect := regexp.MustCompile(`收藏:(?s:(.*?))\n`)
		regOwner := regexp.MustCompile(`作者: \n(?s:(.*?))\n`)

		viewkey := regViewKey.FindAllStringSubmatch(videoUrl, 1)
		addTime := regAddTime.FindAllStringSubmatch(textStr, 1)
		watch := regWatch.FindAllStringSubmatch(textStr, 1)
		collect := regCollect.FindAllStringSubmatch(textStr, 1)
		owner := regOwner.FindAllStringSubmatch(textStr, 1)

		if len(viewkey) > 0 && len(addTime) > 0 && len(watch) > 0 && len(collect) > 0 && len(owner) > 0 && urlOk {

			vi := new(VideoInfo)

			//title = "1"
			strs := strings.Fields(addTime[0][1])

			if len(strs) == 3 {
				switch strs[1] {
				case "分钟":
					duration, _ := time.ParseDuration("-" + strs[0] + "m")
					vi.UpTime = time.Now().Add(duration)
				case "小时":
					duration, _ := time.ParseDuration("-" + strs[0] + "h")
					vi.UpTime = time.Now().Add(duration)
				case "天":
					hourDay, _ := strconv.Atoi(strs[0])
					hourDay = hourDay * 24
					duration, _ := time.ParseDuration("-" + strconv.Itoa(hourDay) + "h")
					vi.UpTime = time.Now().Add(duration)
				}
			}
			vi.Title = title
			vi.ViewKey = viewkey[0][1]
			strs = strings.Fields(watch[0][1])
			vi.Watch, _ = strconv.Atoi(strs[0])
			strs = strings.Fields(collect[0][1])
			if len(strs) > 0 {
				vi.Collect, _ = strconv.Atoi(strs[0])
			}
			vi.Owner = strings.Fields(owner[0][1])[0]
			vMinute := 0
			vSecond := 0
			fmt.Sscanf(selection.Find("span.duration").Text(), "%d:%d\n", &vMinute, &vSecond)
			vi.Vdurat = float64(vMinute) + float64(vSecond)/60.0

			viAll = append(viAll, vi)

			//fmt.Println(vi)
		}

	})
	return
}

func OrgPageSave(dstUrl, proxyUrl, fileName string) {
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

func DownloadMany(viAll []*VideoInfo, numThread int, proxyUrl, savePath string) (failVi, succsVi []*VideoInfo) {
	ch := make(chan int, len(viAll))
	chq := make(chan int, numThread)
	fmt.Print("DownladMany:len([]*VideoInfo)=", len(viAll), "\n")
	var mutex sync.Mutex
	for i, vi := range viAll {
		go func(info *VideoInfo, cnt int) {
			chq <- 1
			info.updateDlAddr(proxyUrl)
			savePath := filepath.Join(savePath, fmt.Sprintf("%s(%s)_%d.ts", info.Title, info.Owner, cnt))
			err := info.Download(savePath, 15, proxyUrl)
			if err != nil {
				mutex.Lock()
				failVi = append(failVi, info)
				mutex.Unlock()
				os.Remove(savePath)
			} else {
				mutex.Lock()
				succsVi = append(succsVi, info)
				mutex.Unlock()
			}

			<-chq
			ch <- 1
		}(vi, i)
		time.Sleep(time.Second * 3)
	}

	for range viAll {
		<-ch
	}
	return
}
