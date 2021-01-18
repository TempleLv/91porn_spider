package score

import (
	"fmt"
	"github.com/yanyiwu/gojieba"
	"io/ioutil"
	"spider91/catch"
	"strconv"
	"strings"
)

type Score struct {
	jieba    *gojieba.Jieba
	keyValue map[string]int
}

func NewScore(keyFile string) *Score {
	mapKv := map[string]int{}
	x := gojieba.NewJieba()
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

	data, err := ioutil.ReadFile(keyFile)
	if err != nil {
		fmt.Println("keyFile read fail:", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		strs := strings.Fields(line)
		if len(strs) == 2 {
			v, err := strconv.Atoi(strs[1])
			if err != nil || v > 100 {
				fmt.Println("wrong key value format!", strs)
				continue
			}
			mapKv[strs[0]] = v
			x.AddWord(strs[0])
		} else {
			fmt.Println("wrong key value format!", strs)
		}

	}

	return &Score{x, mapKv}
}

func (s *Score) Free() {
	s.jieba.Free()
}

func (s *Score) Grade(info catch.VideoInfo) float64 {

	words := s.jieba.Cut(info.Title, true)
	var titleScore, duraScore, viewScore float64
	for _, w := range words {
		titleScore += float64(s.keyValue[w])
	}

	duraScore = 10.0 * info.Vdurat
	if duraScore > 100 {
		duraScore = 100
	}

	viewScore = 0.0

	finalScore := 0.4*titleScore + 0.4*duraScore + viewScore*0.2
	//fmt.Println(finalScore, titleScore, duraScore, 0.35 * titleScore, 0.45*duraScore)

	//for _, vi := range viAll {
	//	fmt.Println(vi.Title)
	//	//words := x.Cut(vi.Title, true)
	//	words := x.Cut(vi.Title, true)
	//	fmt.Println("精确模式:", strings.Join(words, "/"))
	//}
	return finalScore
}
