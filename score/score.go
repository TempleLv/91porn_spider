package score

import (
	"github.com/yanyiwu/gojieba"
	"spider91/catch"
)

func Score(info catch.VideoInfo) float32 {
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

	//for _, vi := range viAll {
	//	fmt.Println(vi.Title)
	//	//words := x.Cut(vi.Title, true)
	//	words := x.Cut(vi.Title, true)
	//	fmt.Println("精确模式:", strings.Join(words, "/"))
	//}
}
