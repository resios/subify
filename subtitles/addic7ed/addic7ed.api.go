package addic7ed

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/matcornic/subify/subtitles/languages"
)

const (
	userAgent = "Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US) AppleWebKit/525.13 (KHTML, like Gecko) Chrome/0.A.B.C Safari/525.13"
	homepage  = "http://www.addic7ed.com"
)

// API entry point
type API struct {
	Name    string
	Aliases []string
}

// New creates a new API for Addic7ed
func New() API {
	return API{
		Name:    "Addic7ed",
		Aliases: []string{"addicted", "addic7ed"},
	}
}

// Download downloads the Addic7ed subtitle for a video
func (s API) Download(videoPath string, language lang.Language) (subtitlePath string, err error) {
	subtitlePath = videoPath[0:len(videoPath)-len(path.Ext(videoPath))] + ".srt"
	// scrape the page and then extract the link
	// narcos.s03e05.web.x264-strife.mkv

	qs := url.Values{}
	qs.Set("search", path.Base(videoPath))
	qs.Set("Submit", "search")

	r, err := http.NewRequest("GET", homepage+"/search.php?"+qs.Encode(), nil)
	if err != nil {
		return
	}

	r.Header.Add("User-Agent", userAgent)
	r.Header.Add("Referer", homepage)

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return
	}

	link, err := extractLink(doc, language, path.Base(videoPath))
	if err != nil {
		return
	}

	err = downloadSubtitle(link, subtitlePath)

	return
}

func downloadSubtitle(link, subtitlePath string) (err error) {
	r, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return
	}

	r.Header.Add("User-Agent", userAgent)
	r.Header.Add("Referer", homepage)

	sub, err := http.DefaultClient.Do(r)
	if err != nil {
		return
	}
	defer sub.Body.Close()

	f, err := os.Create(subtitlePath)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = io.Copy(f, sub.Body)
	if err != nil {
		return
	}

	return
}

type subInfo struct {
	version   string
	downloads int
	sequences int
	completed bool
	matches   bool
	link      string
}

var releasePattern = regexp.MustCompile("Version (.+), ([0-9]+).([0-9])+ MBs")
var statsPattern = regexp.MustCompile("([0-9]+) times edited \u00b7 ([0-9]+) Downloads \u00b7 ([0-9]+) sequences")

func less(p, q subInfo) bool {
	pLink := p.link != ""
	qLink := q.link != ""
	// prefer the one with the link
	switch {
	case pLink && !qLink:
		return false
	case !pLink && qLink:
		return true
	}
	// prefer the one that matches
	switch {
	case p.matches && !q.matches:
		return false
	case !p.matches && q.matches:
		return true
	}

	// prefer complete ones
	switch {
	case p.completed && !q.completed:
		return false
	case !p.completed && q.completed:
		return true
	}

	// prefer most downloaded
	switch {
	case p.downloads < q.downloads:
		return true
	case p.downloads > q.downloads:
		return false
	}

	// prefer the one with more sequences
	switch {
	case p.sequences < q.sequences:
		return true
	case p.sequences > q.sequences:
		return false
	}

	return false
}

func extractLink(doc *goquery.Document, lang lang.Language, video string) (link string, err error) {
	//find all elements that match our language
	byLanguage := doc.Find("td.language").FilterFunction(func(_ int, sel *goquery.Selection) bool {
		return strings.EqualFold(lang.Description, strings.TrimSpace(sel.Text()))
	})

	//log.Printf("Languages: %+v\n", byLanguage)
	main := byLanguage.Closest("table.tabel95")
	// Traverse main and extract the version, downloads, number of sequences, and if it matches the version
	//log.Printf("Main: %+v\n", main)
	var subs []subInfo
	main.Each(func(_ int, sel *goquery.Selection) {
		info := subInfo{}
		news := sel.Find("td.NewsTitle")
		info.version = news.Text()
		// need to check if we have a version
		capture := releasePattern.FindStringSubmatch(info.version)
		if len(capture) < 3 {
			return
		}

		//TODO use regexp to clean this up
		group := strings.Replace(capture[1], "WEB-DL-", "", -1)
		group = strings.Replace(group, "WEBrip.", "", -1)
		info.matches = strings.Contains(strings.ToLower(video), strings.ToLower(group))

		l := sel.Find("td.language")
		info.completed = "Completed" == strings.TrimSpace(l.SiblingsFiltered("td").Find("b").Text())
		u := sel.Find("td.newsDate")
		capture = statsPattern.FindStringSubmatch(strings.TrimSpace(u.Text()))
		if len(capture) > 3 {
			info.downloads, _ = strconv.Atoi(capture[2])
			info.sequences, _ = strconv.Atoi(capture[3])
		}

		info.link = sel.Find("a.buttonDownload").AttrOr("href", "")
		if info.link == "" {
			return
		}

		info.link = homepage + info.link
		log.Printf("Info: %+v\n", info)
		subs = append(subs, info)
	})

	log.Printf("Found %d candidate subtitles\n", len(subs))
	// find max then
	best := subInfo{}
	for _, s := range subs {
		cmp := less(best, s)
		//log.Printf("less(%+v, %+v): %+v\n", best, s, cmp)
		if cmp {
			best = s
		}
	}
	//log.Printf("Best: %+v\n", best)

	if best.link == "" {
		return "", errors.New("Not found")
	}

	return best.link, nil

}

// Upload uploads the subtitle to OpenSubtitles, for the given video
func (s API) Upload(subtitlePath string, langauge lang.Language, videoPath string) error {
	return errors.New("Not yet implemented")
}

// GetName returns the name of the api
func (s API) GetName() string {
	return s.Name
}

// GetAliases returns aliases to identify this API
func (s API) GetAliases() []string {
	return s.Aliases
}
