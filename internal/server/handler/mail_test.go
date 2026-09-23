package handler

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderRecoveryMail(t *testing.T) {
	title, content, err := renderMail("recovery", recoveryMailData{
		AppName:  "0xC0DE",
		Username: "alice",
		Link:     "https://example.com/users/recover/alice.abcdef",
		Lifetime: 15 * time.Minute,
	})
	require.NoError(t, err)

	assert.Equal(t, "[0xC0DE] 계정 복구 안내", title)
	// 제목은 헤더 한 줄이다. 줄바꿈이 남아 있으면 안 된다
	assert.NotContains(t, title, "\n")

	for _, body := range []string{content.Text, content.HTML} {
		assert.Contains(t, body, "alice")
		assert.Contains(t, body, "https://example.com/users/recover/alice.abcdef")
		assert.Contains(t, body, "15m0s")
	}

	assert.Contains(t, content.HTML, "<p>")
	assert.NotContains(t, content.Text, "<p>")
}

// 평문은 text/template, HTML 은 html/template 이라 이스케이프가 서로 다르다.
// 한 파서로 둘 다 읽으면 평문에 &amp; 가 나오거나 HTML 에 태그가 그대로 들어간다.
func TestRenderMailEscaping(t *testing.T) {
	_, content, err := renderMail("recovery", recoveryMailData{
		AppName:  "0xC0DE",
		Username: `<script>alert("x")</script>`,
		Link:     "https://example.com/users/recover/x.y",
		Lifetime: time.Minute,
	})
	require.NoError(t, err)

	assert.Contains(t, content.Text, `<script>alert("x")</script>`)

	assert.NotContains(t, content.HTML, "<script>")
	assert.Contains(t, content.HTML, "&lt;script&gt;")
}

// 링크는 그대로 살아 있어야 한다. html/template 이 URL 을 지워버리면 메일이 쓸모없어진다
func TestRenderMailKeepsLink(t *testing.T) {
	link := "https://example.com/users/recover/alice.aB3-_xyz"

	_, content, err := renderMail("recovery", recoveryMailData{
		AppName:  "0xC0DE",
		Username: "alice",
		Link:     link,
		Lifetime: time.Minute,
	})
	require.NoError(t, err)

	assert.Contains(t, content.HTML, `href="`+link+`"`)
	assert.NotContains(t, content.HTML, "ZgotmplZ") // html/template 이 URL 을 막았을 때 넣는 표식
}

// 세 파일이 다 있어야 한 통이 된다
func TestRenderMailMissingTemplate(t *testing.T) {
	_, _, err := renderMail("no-such-mail", nil)
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "no-such-mail"), err.Error())
}
