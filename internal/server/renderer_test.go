package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 템플릿 문법이 깨지면 서버가 뜰 때까지 모른다. 파싱만이라도 미리 해 본다.
func TestNewRenderer(t *testing.T) {
	// dev 빌드의 assets 는 작업 디렉터리 기준으로 파일을 읽는다
	t.Chdir("../..")

	tmpl, err := NewRenderer()
	require.NoError(t, err)

	names := map[string]bool{}
	for _, t := range tmpl.Templates() {
		names[t.Name()] = true
	}

	for _, name := range []string{
		"login.html",
		"register.html",
		"users/passkeys.html",
		"users/recover.html",
		"users/recover-confirm.html",
		"users/settings.html",
	} {
		assert.True(t, names[name], "template %q is missing", name)
	}
}
