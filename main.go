package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bluele/slack"
)

type event struct {
	ID    string
	Date  string
	Title string
	URL   string
}

func main() {
	dir := path.Dir(os.Args[0])
	filePath := path.Join(dir, "loading.txt")

	lines := fromFile(filePath)
	eventList := crawl(lines[0])
	if 0 < len(eventList) {
		writeFile(filePath, eventList[0].ID)
		notifySlack(eventList)
	}
}

func crawl(lastLoadID string) []event {
	const URL = "https://techplay.jp/event/search"
	const PERPAGE = 15
	const MAXPAGE = 6

	page := 1
	eventList := make([]event, 0, PERPAGE*MAXPAGE)

	for {
		values := url.Values{}
		values.Add("pref", "13,14")
		values.Add("sort", "created_desc")
		values.Add("page", strconv.Itoa(page))

		events := parseHTML(URL + "?" + values.Encode())
		for _, event := range events {
			if event.ID == lastLoadID {
				break
			}

			eventList = append(eventList, event)
		}

		if len(eventList) != page*PERPAGE || page == MAXPAGE {
			break
		}
		page++
	}

	return eventList
}

func parseHTML(url string) []event {
	doc, _ := goquery.NewDocument(url)
	eventListDocument := doc.Find("article > div.eventlist")
	eventLength := eventListDocument.Length()
	events := make([]event, eventLength)

	eventListDocument.Each(func(i int, s *goquery.Selection) {
		date := parseDateDocument(s.Find("div.date"))
		title, href := parseTitleDocument(s.Find("div.title"))

		re := regexp.MustCompile("https://eventdots.jp/event/(\\d+)")
		id := re.ReplaceAllString(href, "$1")

		events[i] = event{ID: id, Date: date, Title: title, URL: href}
	})
	return events
}

func parseDateDocument(s *goquery.Selection) string {
	year := s.Find("div.year").Text()
	day := ""
	time := ""

	days := s.Find("div.day")
	if 1 < days.Length() {
		tmp := days.First()
		time = tmp.Find("span.time").Text()

		html, _ := tmp.Html()
		re := regexp.MustCompile("<span.*</span>")
		day = strings.TrimSpace(re.ReplaceAllString(html, ""))
	} else {
		day = days.Text()
		times := s.Find("div.time span")
		startTime := times.First().Text()
		endTime := times.Last().Text()
		time = startTime + "〜" + endTime
	}
	return fmt.Sprintf("%s/%s %s", year, day, time)
}

func parseTitleDocument(s *goquery.Selection) (title string, href string) {
	a := s.Find("h3 > a")
	title = a.Text()
	href, _ = a.Attr("href")
	return
}

func fromFile(filePath string) []string {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "File %s could not read: %v\n", filePath, err)
		os.Exit(1)
	}

	defer f.Close()

	lines := make([]string, 0, 1)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if serr := scanner.Err(); serr != nil {
		fmt.Fprintf(os.Stderr, "File %s scan error: %v\n", filePath, err)
	}

	return lines
}

func writeFile(filePath string, id string) {
	content := []byte(id)
	ioutil.WriteFile(filePath, content, os.ModePerm)
}

func notifySlack(eventList []event) {
	webhook_url := os.Getenv("SLACK_WEBHOOK_URL")
	channel := os.Getenv("SLACK_WEBHOOK_CHANNEL")
	if webhook_url == "" || channel == "" {
		return
	}

	api := slack.NewWebHook(webhook_url)
	r := regexp.MustCompile(`Pepper`)
	for _, event := range eventList {
		if !r.MatchString(event.Title) {
			message := fmt.Sprintf("日時: %s\nイベント名: %s\nURL: %s", event.Date, event.Title, event.URL)
			err := api.PostMessage(&slack.WebHookPostPayload{
				Text:    message,
				Channel: channel,
			})
			if err != nil {
				panic(err)
			}
		}
	}
}
