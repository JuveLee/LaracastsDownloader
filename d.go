package main

import (
	"flag"
	"log"
	"os"
)

var (
	debugMode  bool
	resolution string
	targetPath string
	proxy      string
)

func main() {

	listURL := flag.String("l", "", "list https://laracasts.com/series/charting-and-you")
	videoURL := flag.String("v", "", "url https://laracasts.com/series/charting-and-you/episodes/1")
	flagPath := flag.String("s", ".", "save path /tmp")
	flagProxy := flag.String("p", "", "Proxy eg http://8.8.8.8:8080")
	re := flag.String("r", "hd", "resolution in hd or sd")
	debug := flag.Bool("d", false, "debug mode")

	flag.Parse()
	var cmd string = flag.Arg(0)
	debugMode = *debug
	resolution = *re
	proxy = *flagProxy
	targetPath = normalize(*flagPath, "path")

	if targetPath != `.` {
		err := os.MkdirAll(targetPath, 0777)
		if err != nil {
			panic(err)
		}
	}

	if len(*videoURL) > 0 {
		log.Print("video")
		downloadFromV(*videoURL, ``)
	}

	if len(*listURL) > 0 {
		log.Print("list")
		getList(*listURL)
	}

	if len(cmd) != 0 {
		log.Print("video")
		downloadFromV(cmd, ``)
	}

	//fmt.Println(getHTML(url))
	// p, err := goquery.ParseUrl(url)
	// if err != nil {
	// 	panic(err)
	// } else {
	// 	pTitle := p.Find("title").Text() //直接提取title的内容
	// 	fmt.Println(pTitle)
	// }
}
