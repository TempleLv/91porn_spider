package catch

import (
	"context"
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
		//chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3830.0 Safari/537.36"),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, _ = context.WithTimeout(ctx, time.Second*15)

	htmlText := ""
	fullUrl := "https://www.91porn.com/view_video.php?viewkey=" + v.ViewKey
	if err = chromedp.Run(ctx, sourHtml(fullUrl, "#player_one_html5_api > source", &htmlText)); err != nil {
		fmt.Println(err)
		return
	}
	regAddr := regexp.MustCompile(`<source src="(?s:(.*?))" type="`)
	dlAddr := regAddr.FindAllStringSubmatch(htmlText, 1)
	if len(dlAddr) > 0 {
		v.DlAddr = dlAddr[0][1]
	}

	return
}

func (v VideoInfo) Download(savePath string, numThread int, proxy string) (err error) {

	if len(v.DlAddr) > 0 {
		//strCmd := fmt.Sprintf(" -p \"%s\" -t %d -w -o %s \"%s\"", proxy, numThread, savePath, v.DlAddr)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2+time.Second*30)
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
		chromedp.Navigate(urlstr),
		//chromedp.WaitVisible(sel),
		//chromedp.Text("source src", html),
		//chromedp.ActionFunc(func(ctx context.Context) error {
		//	return nil
		//}),
		chromedp.OuterHTML(sel, html),
	}
}

func PageCrawl(dstUrl, proxyUrl string) (viAll []*VideoInfo) {
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
			vi.Collect, _ = strconv.Atoi(strs[0])
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

func DownloadMany(viAll []*VideoInfo, numThread int, proxyUrl, savePath string) (failVi []*VideoInfo) {
	ch := make(chan int, len(viAll))
	chq := make(chan int, numThread)
	fmt.Print("DownladMany:len([]*VideoInfo)=", len(viAll), "\n")
	for _, vi := range viAll {
		go func(info *VideoInfo) {
			chq <- 1
			info.updateDlAddr(proxyUrl)
			savePath := filepath.Join(savePath, fmt.Sprintf("%s(%s).ts", info.Title, info.Owner))
			err := info.Download(savePath, 10, proxyUrl)
			if err != nil {
				failVi = append(failVi, info)
			}
			<-chq
			ch <- 1
		}(vi)
		time.Sleep(time.Second * 3)
	}

	for range viAll {
		<-ch
	}
	return
}
