package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"spider91/catch"
	"spider91/doneDB"
	"spider91/score"
	"strconv"
	"strings"
	"time"
)

func weeklyFunc() func() {
	return func() {

		log.Println("Start weekly organize!!")
		savePath := time.Now().Format("./save/weekly_060102")

		path, _ := filepath.Abs(savePath)

		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				log.Println("savePath create failed!", err)
				return
			}
		}
		fi, _ := ioutil.ReadDir("./save")

		for _, f := range fi {
			if f.IsDir() && strings.Contains(f.Name(), "daily") {
				os.Rename(filepath.Join("./save", f.Name()), filepath.Join(path, f.Name()))
			}
		}
	}
}

func dailyFunc(proxyUrls []string) func() {
	proxyUrls = append([]string{}, proxyUrls...)
	return func() {
		log.Println("Start daily Download!!")
		var viAll []*catch.VideoInfo
		s := score.NewScore("./score/wordValue.txt")
		defer s.Free()
	CRAWL:
		for i := 1; i < 50; i++ {
			var vis []*catch.VideoInfo
			for _, pu := range proxyUrls {
				vis = catch.PageCrawl("http://91porn.com/v.php?next=watch&page="+strconv.Itoa(i), pu)
				if len(vis) > 0 {
					break
				}
			}

			for _, vi := range vis {
				if time.Now().Sub(vi.UpTime) < time.Hour*24+time.Minute*10 {
					viAll = append(viAll, vi)
				} else {
					break CRAWL
				}
			}
		}

		if len(viAll) > 0 {
			ddb, err := doneDB.OpenVDB("./save/videoDB.db")
			if err != nil {
				log.Println("videoDB.db open fail!!!", err)
				return
			}
			defer ddb.Close()

			viAll = ddb.DelRepeat(viAll)
			s.GradeSort(viAll)
			length := int(math.Min(40, float64(len(viAll))))
			pickVi := append(viAll[:length], ddb.GetUD()...)
			savePath := time.Now().Format("./save/daily_060102")

			path, _ := filepath.Abs(savePath)

			_, err = os.Stat(path)
			if os.IsNotExist(err) {
				if err = os.MkdirAll(path, os.ModePerm); err != nil {
					log.Println("savePath create failed!", err)
					return
				}
			}

			failVi := pickVi
			for _, pu := range proxyUrls {
				failVi = catch.DownloadMany(failVi, 5, pu, path)
				if len(failVi) == 0 {
					break
				} else {
					log.Printf("proxy:%s left %d items\n", pu, len(failVi))
				}
			}
			ddb.AddDone(pickVi)
			ddb.UpdateUD(failVi)
			log.Printf("Download total:%d, success %d, fail %d.\n", len(pickVi), len(pickVi)-len(failVi), len(failVi))
			for _, vi := range failVi {
				log.Println("Download Fail!", vi.Title, vi.ViewKey)
			}
		} else {
			log.Println("No page was crawled!!!")
		}

	}
}

func main() {

	//ddb, err1 := doneDB.OpenVDB("./save/videoDB.db")
	//if err1 != nil {
	//	panic(err1)
	//}
	//defer ddb.Close()
	//
	//viAll := catch.PageCrawl("http://91porn.com/index.php", "")
	////ddb.AddDone(viAll)
	//ddb.UpdateUD(viAll)
	////viAll = ddb.DelRepeat(viAll)
	//viAll = ddb.GetUD()
	//
	//return

	proxyUrl := ""
	pageUrl := ""
	savePath := ""
	threadNum := 5
	cpage := false

	flag.StringVar(&proxyUrl, "p", "", "proxy")
	flag.StringVar(&pageUrl, "u", "http://91porn.com/index.php", "page to crawl")
	flag.StringVar(&savePath, "o", "./save", "path to output")
	flag.IntVar(&threadNum, "t", 5, "threadcount")
	flag.BoolVar(&cpage, "c", false, "crawl whole page")

	flag.Parse()

	if cpage == true {
		path, _ := filepath.Abs(savePath)

		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				fmt.Println(err)
			}
		}

		viAll := catch.PageCrawl(pageUrl, proxyUrl)

		catch.DownloadMany(viAll, threadNum, proxyUrl, path)

		return
	}

	path, _ := filepath.Abs("./save/")

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			log.Println("log path create failed!", err)
			return
		}
	}

	logFile, err := os.OpenFile("./save/spider91.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	logW := io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(logW)

	log.Println("spider91 startup!")

	c := cron.New(cron.WithSeconds())

	c.AddFunc("00 00 04 * * 6", weeklyFunc())

	proxyUrls := []string{
		"",
		"socks5://192.168.3.254:10808",
	}
	c.AddFunc("00 00 03 * * *", dailyFunc(proxyUrls))

	c.Start()
	defer c.Stop()

	select {}
}
