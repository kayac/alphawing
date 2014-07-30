package models

import (
	"io/ioutil"

	"github.com/google/go-github/github"
)

func RenderMarkdown(md string) (string, error) {
	client := github.NewClient(nil)

	html, _, err := client.Markdown(md, nil)
	if err != nil {
		return "", err
	}

	return html, nil
}

func GenerateApiDocumentHtml(srcPath string) (string, error) {
	md, err := ioutil.ReadFile(srcPath)
	if err != nil {
		panic(err)
	}

	html, err := RenderMarkdown(string(md))
	if err != nil {
		return "", err
	}

	html = "<div class=\"github-markdown\">\n" + html + "<!-- /.github-markdown --></div>\n"
	html = "<section class=\"api-document\">\n" + html + "<!-- /.api-document --></section>\n"
	html = "{{template \"header.html\" .}}\n" + html + "{{template \"footer.html\" .}}"
	html = "{{set . \"title\" \"API Document\"}}\n" + html

	return html, nil
}
