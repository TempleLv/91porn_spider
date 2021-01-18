package main

import (
	"fmt"
	cron "github.com/robfig/cron/v3"
	"github.com/yanyiwu/gojieba"
	"log"
	"spider91/catch"
	"strconv"
	"strings"
	"time"
)

func main() {
	//log.Println("starting.....")
	//c := cron.New(cron.WithSeconds())
	//
	//c.AddFunc("* * * * * *", func() {
	//	log.Println(time.Now())
	//})
	//
	//c.Start()
	//defer c.Stop()
	//
	//select {
	//
	//}

	//proxyUrl := ""
	//pageUrl := ""
	//savePath := ""
	//threadNum := 5
	//
	//flag.StringVar(&proxyUrl, "p", "", "proxy")
	//flag.StringVar(&pageUrl, "u", "http://91porn.com/index.php", "page to crawl")
	//flag.StringVar(&savePath, "o", "./save", "path to output")
	//flag.IntVar(&threadNum, "t", 5, "thradcount")
	//
	//flag.Parse()
	//
	//path, _ := filepath.Abs(savePath)
	//
	//_, err := os.Stat(path)
	//if os.IsNotExist(err) {
	//	if err = os.MkdirAll(path, os.ModePerm); err != nil {
	//		fmt.Println(err)
	//	}
	//}
	//
	//viAll := pageCrawl(pageUrl, proxyUrl)
	//
	//DownladMany(viAll, threadNum, proxyUrl, path)
	//
	//return

	log.Println("starting.....")
	c := cron.New(cron.WithSeconds())

	c.AddFunc("*/5 * * * * *", func() {
		var viAll []*catch.VideoInfo
	ALL:
		for i := 1; i < 50; i++ {
			vis := catch.PageCrawl("http://91porn.com/v.php?next=watch&page="+strconv.Itoa(i), "http://192.168.4.66:10808")
			for _, vi := range vis {
				if time.Now().Sub(vi.UpTime) < time.Hour*24+time.Minute*10 {
					viAll = append(viAll, vi)
				} else {
					break ALL
				}
			}

		}

		x := gojieba.NewJieba()
		defer x.Free()
		x.AddWord("真舒服")
		x.AddWord("草死")
		x.AddWord("网红")
		x.AddWord("小女友")
		x.AddWord("大屁股")
		x.AddWord("跳蛋")
		x.AddWord("女室友")
		x.AddWord("D奶")
		x.AddWord("d奶")
		x.AddWord("E奶")
		x.AddWord("e奶")
		x.AddWord("F奶")
		x.AddWord("f奶")
		x.AddWord("小骚货")
		x.AddWord("魔都")
		x.AddWord("微露脸")
		x.AddWord("干进去")
		x.AddWord("肉臀")
		x.AddWord("学霸")
		x.AddWord("小母狗")
		x.AddWord("高跟")
		x.AddWord("00后")

		for _, vi := range viAll {
			fmt.Println(vi.Title)
			//words := x.Cut(vi.Title, true)
			words := x.Cut(vi.Title, true)
			fmt.Println("精确模式:", strings.Join(words, "/"))
		}

	})

	c.Start()
	defer c.Stop()

	select {}

}
