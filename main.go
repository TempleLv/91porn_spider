package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
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

	logFile, err := os.OpenFile("spider91.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	logW := io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(logW)

	log.Println("uptime:", time.Now().Format("2006-01-02 15:04:05"))

	savePath := time.Now().Format("./save/060102")
	fmt.Println(savePath)

	c := cron.New(cron.WithSeconds())

	c.AddFunc("0 50 15 * * *", func() {
		log.Println("Start Download!!")
		var viAll []*catch.VideoInfo
		s := score.NewScore("./score/wordValue.txt")
		defer s.Free()
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

		s.GradeSort(viAll)
		length := int(math.Min(40, float64(len(viAll))))
		pickVi := viAll[:length]
		savePath := time.Now().Format("./save/060102")

		path, _ := filepath.Abs(savePath)

		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				log.Println("savePath create failed!", err)
				return
			}
		}

		failVi := catch.DownloadMany(pickVi, 5, "", savePath)

		for _, vi := range failVi {
			log.Println("Download Fail!", vi.Title, vi.ViewKey)
		}
	})

	c.Start()
	defer c.Stop()

	select {}

}
