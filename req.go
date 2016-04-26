package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"gopkg.in/cheggaaa/pb.v1"

	"github.com/PuerkitoBio/goquery"
	"github.com/parnurzeal/gorequest"
)

var (
	// prefix of laracasts
	lPrefix = "https://laracasts.com"
	// prefix of video download url
	vPrefix = "https:"

	replaceChar = "_"
)

//getHTML simply return the html of a url
func getHTML(url string) (html string) {
	request := gorequest.New().SetDebug(debugMode)
	if len(proxy) != 0 {
		request.Proxy(proxy)
	}
	_, body, _ := request.Get(url).
		Set("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,`+
		`image/webp,*/*;q=0.8`).
		Set("Pragma", "no-cache").
		Set("Referer", "https://laracasts.com/").
		Set("Connection", "close").
		Set("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_4)`+
		`AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.86 Safari/537.36`).
		End()
	return body
}

// getList deals with list url
func getList(listURL string) {
	doc := parseURL(listURL)
	var videoList []map[string]string
	doc.Find(".Lesson-List--numbered").
		Find("li").
		Find("a").
		Has("span").
		Each(func(i int, s *goquery.Selection) {
		tempMap := make(map[string]string)
		origin := strings.Split(strings.TrimSpace(s.Text()), "\n")[0]
		tempMap[`name`] = origin
		tempMap[`key`] = strconv.Itoa(i + 1)
		origin, _ = s.Attr("href")
		tempMap[`url`] = normalize(origin, "url")
		videoList = append(videoList, tempMap)
	})
	total := len(videoList)
	log.Printf("Total %d", total)
	//add 0,i counts the digit of length
	var i int
	for ; ; i++ {
		if total/int(math.Pow10(i)) == 0 {
			break
		}
	}

	for _, v := range videoList {
		key, _ := strconv.Atoi(v[`key`])
		downloadFromV(v[`url`], fmt.Sprintf("%0*d", i, key)+`.`+v[`name`])
	}
}

// normalize the name or the path
func normalize(s, nType string) string {

	switch nType {
	case "name":
		if len(s) == 0 {
			return ``
		}
		return strings.Replace(strings.Replace(s, `/`, replaceChar, -1), ` `, replaceChar, -1)
	case "url":
		if !strings.HasPrefix(s, `http`) {
			return lPrefix + s
		}
		return s
	case "v":
		if !strings.HasPrefix(s, `http`) {
			return vPrefix + s
		}
		return s

	case "path":
		return strings.TrimRight(s, string(os.PathSeparator))
	default:
		panic("Unknown type")
	}

}

// downloadFromV downloads a single video
func downloadFromV(videoURL, name string) {
	log.Printf("videoURL %s", videoURL)

	doc := parseURL(videoURL)
	videoURL, videoType := getVideoURL(doc)
	if len(name) == 0 {
		//try to give a name
		name = doc.Find("title").Text()
	}
	// normalize
	name = normalize(name, "name") + videoType
	downloadVideo(videoURL, name)
}

// getVideoURL
func getVideoURL(doc *goquery.Document) (string, string) {
	var v string
	// get right rs
	doc.Find(`source`).
		Each(func(i int, s *goquery.Selection) {
		if re, _ := s.Attr("data-quality"); re == strings.ToUpper(resolution) {
			v, _ = s.Attr(`src`)
		}
	})
	if v == `` {
		log.Panic("no video")
	}
	// get ext
	videoType := path.Ext(strings.Split(v, `?`)[0])

	return normalize(v, "v"), videoType
}

func downloadVideo(downloadURL, fileName string) (err error) {
	log.Printf("fileName %s", fileName)
	log.Printf("downloadURL %s", downloadURL)
	if len(proxy) != 0 {
		proxyURL, _ := url.Parse(proxy)
		http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", downloadURL, nil)
	req.Header.Add("Referer", `https://laracasts.com/`)
	req.Header.Add("Pragma", `no-cache`)
	req.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_4)`+
		`AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.86 Safari/537.36`)

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		panic(err)
	}

	contentLength := resp.ContentLength
	log.Print(contentLength)
	bar := pb.New(int(contentLength)).SetUnits(pb.U_BYTES)

	//Create the file
	targetName := targetPath + string(os.PathSeparator) + fileName
	log.Printf("targetName %s", targetName)
	bar.Start()
	out, err := os.Create(targetName)
	if err != nil {
		return err
	}
	defer out.Close()

	writer := io.MultiWriter(out, bar)

	// Writer the body to file
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return err
	}
	bar.FinishPrint(fileName + " Finished!")
	return
}

func parseURL(u string) *goquery.Document {
	stringHTML := getHTML(u)
	return parseHTML(stringHTML)
}

func parseHTML(stringHTML string) *goquery.Document {
	r := strings.NewReader(stringHTML)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}
