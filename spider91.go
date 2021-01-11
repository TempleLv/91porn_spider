package main

import (
	"fmt"
	"net/http"
	"regexp"
	"src/github.com/PuerkitoBio/goquery"
)

func pageCrawl(url string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	client := http.Client{}

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
			fmt.Println(title, viewkey[0][1], addTime[0][1])
		}

	})
}

func main() {
	pageCrawl("http://91porn.com/index.php")
}
