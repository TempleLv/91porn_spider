package score

import (
	"fmt"
	"github.com/yanyiwu/gojieba"
	"io/ioutil"
	"sort"
	"spider91/catch"
	"strconv"
	"strings"
)

type Score struct {
	jieba    *gojieba.Jieba
	keyValue map[string]int
	ownValue map[string]int
}

func NewScore(keyFile, ownFile string) *Score {
	mapKv := map[string]int{}
	mapOv := map[string]int{}
	x := gojieba.NewJieba()

	data, err := ioutil.ReadFile(keyFile)
	if err != nil {
		fmt.Println("keyFile read fail:", err)
	}

	owndata, err := ioutil.ReadFile(ownFile)
	if err != nil {
		fmt.Println("ownFile read fail:", err)
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

	for _, line := range strings.Split(string(owndata), "\n") {
		strs := strings.Fields(line)
		if len(strs) == 2 {
			v, err := strconv.Atoi(strs[1])
			if err != nil || v > 100 {
				fmt.Println("wrong own value format!", strs)
				continue
			}
			mapOv[strs[0]] = v
		} else {
			fmt.Println("wrong own value format!", strs)
		}
	}

	return &Score{x, mapKv, mapOv}
}

func (s *Score) Free() {
	s.jieba.Free()
}

func (s *Score) Grade(info *catch.VideoInfo) float64 {

	words := s.jieba.Cut(info.Title, true)
	var titleScore, duraScore, ownScore float64
	for _, w := range words {
		titleScore += float64(s.keyValue[w])
	}

	duraScore = 10.0 * info.Vdurat
	if duraScore > 100 {
		duraScore = 100
	}

	ownScore = float64(s.ownValue[info.Owner])

	finalScore := 0.4*titleScore + 0.4*duraScore + 0.2*ownScore

	return finalScore
}

func (s *Score) GradeSort(vis []*catch.VideoInfo) {
	for _, vi := range vis {
		vi.Score = s.Grade(vi)
	}
	sort.Sort(sort.Reverse(catch.ViSlice(vis)))
}
