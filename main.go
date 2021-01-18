package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"spider91/catch"
	"spider91/score"
	"strconv"
	"time"
)

func main() {

	//proxyUrl := ""
	//pageUrl := ""
	//savePath := ""
	//threadNum := 5
	//
	//flag.StringVar(&proxyUrl, "p", "", "proxy")
	//flag.StringVar(&pageUrl, "u", "http://91porn.com/index.php", "page to crawl")
	//flag.StringVar(&savePath, "o", "./save", "path to output")
	//flag.IntVar(&threadNum, "t", 5, "threadcount")
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
	//viAll := catch.PageCrawl(pageUrl, proxyUrl)
	//
	//catch.DownladMany(viAll, threadNum, proxyUrl, path)
	//
	//return

	log.Println("starting.....")

	c := cron.New(cron.WithSeconds())

	c.AddFunc("*/5 * * * * *", func() {
		var viAll []*catch.VideoInfo
		s := score.NewScore("./score/wordValue.txt")
		defer s.Free()
	ALL:
		for i := 1; i < 50; i++ {
			vis := catch.PageCrawl("http://91porn.com/v.php?next=watch&page="+strconv.Itoa(i), "")
			for _, vi := range vis {
				if time.Now().Sub(vi.UpTime) < time.Hour*24+time.Minute*10 {
					viAll = append(viAll, vi)
				} else {
					break ALL
				}
			}

		}
		for _, vi := range viAll {
			fmt.Println(vi.Title)
			s.Grade(*vi)
		}
	})

	c.Start()
	defer c.Stop()

	select {}

}
