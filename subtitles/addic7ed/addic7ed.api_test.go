package addic7ed

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/matcornic/subify/subtitles/languages"
)

func Test_extractLink(t *testing.T) {

	type args struct {
		doc   *goquery.Document
		lang  lang.Language
		video string
	}
	tests := []struct {
		name     string
		args     args
		wantLink string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{name: "T1", wantErr: false, wantLink: "http://www.addic7ed.com/original/127035/0", args: args{
			doc:   loadDocument(t, "subtitles/addic7ed/test_search.html"),
			lang:  *lang.Languages.GetLanguage("eng"),
			video: "narcos.s03e05.web.x264-strife.mkv",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLink, err := extractLink(tt.args.doc, tt.args.lang, tt.args.video)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLink != tt.wantLink {
				t.Errorf("extractLink() = %v, want %v", gotLink, tt.wantLink)
			}
		})
	}
}

func loadDocument(t *testing.T, path string) (doc *goquery.Document) {
	f, err := os.Open(filepath.Join("testdata", "test_search.html"))
	if err != nil {
		t.Fatalf("Failed to open file: %+v", err)
	}

	doc, err = goquery.NewDocumentFromReader(f)
	if err != nil {
		t.Fatalf("Failed to parse file: %+v", err)
	}
	return
}
