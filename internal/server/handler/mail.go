package handler

import (
	"bytes"
	"embed"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"

	"github.com/cockroachdb/errors"

	"github.com/bluemir/0xC0DE/internal/mail"
)

//go:embed mail-templates
var mailTemplateFS embed.FS

// 메일 한 통은 이름이 같은 세 파일로 이뤄진다. recovery.title, recovery.txt, recovery.html
//
// 평문과 HTML 을 다른 파서로 읽는다. html/template 은 값을 끼워 넣을 때 HTML 로 이스케이프하고
// text/template 은 그대로 둔다. 한 파서로 둘 다 처리하면 평문 본문에 &amp; 가 튀어나오거나,
// 반대로 HTML 본문에 사용자가 적은 <script> 가 그대로 들어간다.
var (
	mailTitleTemplates = texttemplate.Must(
		texttemplate.ParseFS(mailTemplateFS, "mail-templates/*.title"))
	mailTextTemplates = texttemplate.Must(
		texttemplate.ParseFS(mailTemplateFS, "mail-templates/*.txt"))
	mailHTMLTemplates = htmltemplate.Must(
		htmltemplate.ParseFS(mailTemplateFS, "mail-templates/*.html"))
)

// renderMail 은 name 에 해당하는 세 파일을 렌더해 제목과 본문을 만든다.
func renderMail(name string, data any) (string, mail.Content, error) {
	title := &bytes.Buffer{}
	if err := mailTitleTemplates.ExecuteTemplate(title, name+".title", data); err != nil {
		return "", mail.Content{}, errors.WithStack(err)
	}

	text := &bytes.Buffer{}
	if err := mailTextTemplates.ExecuteTemplate(text, name+".txt", data); err != nil {
		return "", mail.Content{}, errors.WithStack(err)
	}

	html := &bytes.Buffer{}
	if err := mailHTMLTemplates.ExecuteTemplate(html, name+".html", data); err != nil {
		return "", mail.Content{}, errors.WithStack(err)
	}

	// 제목은 한 줄이다. 파일 끝의 줄바꿈이 헤더에 섞이지 않게 떼어낸다
	return strings.TrimSpace(title.String()), mail.Content{
		Text: text.String(),
		HTML: html.String(),
	}, nil
}
