package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Hymn struct {
	SongNumber  int    `json:"song_number"`
	Name        string `json:"name"`
	Lyric       string `json:"lyric"`
	Category    string `json:"cat_top"`
	SubCategory string `json:"cat_sub"`
}

var regexps = map[string]string{
	"a": `a|á|à|ả|ã|ạ|ă|ắ|ằ|ẳ|ẵ|ặ|â|ấ|ầ|ẩ|ẫ|ậ`,
	"e": `e|é|è|ẽ|ẻ|ẹ|ê|ế|ề|ễ|ể|ệ`,
	"o": `o|ó|ò|õ|ỏ|ọ|ô|ố|ồ|ỗ|ổ|ộ|ơ|ớ|ờ|ỡ|ở|ợ`,
	"u": `u|ú|ù|ũ|ủ|ụ|ư|ứ|ừ|ữ|ử|ự`,
	"d": `đ`,
	"-": `\s`,
}

func getFileName(songNumber int, songName string) string {
	fileName := strconv.Itoa(songNumber)

	songName = strings.ToLower(songName)

	for repl, r := range regexps {
		re := regexp.MustCompile(r)
		songName = re.ReplaceAllString(songName, repl)
	}

	return fileName + "-" + songName
}

func capitalize(str string) string {
	return strings.Title(strings.ToLower(str))
}

func writeMarkdownFile(fileName string, data Hymn) error {
	f, err := os.Create("./content/hymn/" + fileName + ".md")
	if err != nil {
		return err
	}
	defer f.Close()

	_, _ = f.WriteString("---\n")
	_, _ = f.WriteString(fmt.Sprintf("title: %d. %s\n", data.SongNumber, capitalize(data.Name)))
	_, _ = f.WriteString(fmt.Sprintf("weight: %d\n", data.SongNumber))
	_, _ = f.WriteString(fmt.Sprintf("categories: %s\n", capitalize(data.SubCategory)))
	_, _ = f.WriteString(fmt.Sprintf("draft: %t\n", false))
	_, _ = f.WriteString("---\n")
	_, _ = f.WriteString(data.Lyric)

	f.Sync()

	return nil
}

func crawl(songNumber int, resultChan chan interface{}) {
	targetURL := "http://wiki.cdnvn.com/doc-sach/TC/thanh-ca"
	songURL := fmt.Sprintf("%s/%d?format=json", targetURL, songNumber)
	resp, err := http.Get(songURL)

	if resp.StatusCode != 200 {
		resultChan <- "404 not found"
		return
	}

	if err != nil {
		resultChan <- err
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resultChan <- err
		return
	}

	hymn := Hymn{}
	hymn.SongNumber = songNumber
	err = json.Unmarshal(body, &hymn)
	if err != nil {
		resultChan <- err
		return
	}

	fileName := getFileName(songNumber, hymn.Name)
	writeMarkdownFile(fileName, hymn)

	resultChan <- fileName
}

func main() {
	fmt.Println("Start crawling")

	maxSongs := 903
	resultChan := make(chan interface{}, maxSongs)

	for i := 1; i <= maxSongs; i++ {
		go crawl(i, resultChan)
	}

	for i := 0; i < maxSongs; i++ {
		fmt.Println("Recevied:", <-resultChan)
	}

	fmt.Println("Finish")
}
