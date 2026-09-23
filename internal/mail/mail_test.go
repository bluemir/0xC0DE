package mail

import (
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	netmail "net/mail"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parse 는 build 가 만든 메시지를 다시 읽어 헤더와 본문을 돌려준다.
// 문자열을 그대로 비교하면 boundary 와 Date 때문에 깨져서, 파싱해서 본다.
func parse(t *testing.T, raw []byte) (netmail.Header, string) {
	t.Helper()

	msg, err := netmail.ReadMessage(bytes.NewReader(raw))
	require.NoError(t, err)

	body, err := io.ReadAll(msg.Body)
	require.NoError(t, err)

	return msg.Header, string(body)
}

func decodeQuotedPrintable(t *testing.T, s string) string {
	t.Helper()

	buf, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(s)))
	require.NoError(t, err)

	return string(buf)
}

func TestBuildPlainText(t *testing.T) {
	raw, err := build("보내는이 <from@example.com>", []string{"to@example.com"}, "제목", Content{
		Text: "본문입니다",
	})
	require.NoError(t, err)

	header, body := parse(t, raw)

	assert.Equal(t, "보내는이 <from@example.com>", header.Get("From"))
	assert.Equal(t, "to@example.com", header.Get("To"))
	assert.Equal(t, "text/plain; charset=utf-8", header.Get("Content-Type"))
	assert.Equal(t, "quoted-printable", header.Get("Content-Transfer-Encoding"))
	assert.NotEmpty(t, header.Get("Date"))
	assert.Equal(t, "본문입니다", decodeQuotedPrintable(t, body))
}

func TestBuildEncodesKoreanSubject(t *testing.T) {
	raw, err := build("from@example.com", []string{"to@example.com"}, "비밀번호 재설정 안내", Content{
		Text: "본문",
	})
	require.NoError(t, err)

	header, _ := parse(t, raw)

	// 인코딩되지 않고 날것으로 나가면 클라이언트에서 깨진다
	assert.NotEqual(t, "비밀번호 재설정 안내", header.Get("Subject"))

	decoded, err := new(mime.WordDecoder).DecodeHeader(header.Get("Subject"))
	require.NoError(t, err)
	assert.Equal(t, "비밀번호 재설정 안내", decoded)
}

// 제목에 개행을 넣어 헤더를 하나 더 만들어내는 인젝션을 막는지 본다
func TestBuildSubjectHeaderInjection(t *testing.T) {
	raw, err := build("from@example.com", []string{"to@example.com"},
		"hello\r\nBcc: attacker@example.com", Content{Text: "본문"})
	require.NoError(t, err)

	header, _ := parse(t, raw)

	assert.Empty(t, header.Get("Bcc"))
	assert.NotContains(t, header.Get("Subject"), "\n")
}

func TestBuildMultipleRecipients(t *testing.T) {
	raw, err := build("from@example.com",
		[]string{"a@example.com", "받는이 <b@example.com>"},
		"제목", Content{Text: "본문"})
	require.NoError(t, err)

	header, _ := parse(t, raw)

	addresses, err := header.AddressList("To")
	require.NoError(t, err)
	require.Len(t, addresses, 2)
	assert.Equal(t, "a@example.com", addresses[0].Address)
	assert.Equal(t, "b@example.com", addresses[1].Address)
	assert.Equal(t, "받는이", addresses[1].Name)
}

func TestBuildMultipart(t *testing.T) {
	raw, err := build("from@example.com", []string{"to@example.com"}, "제목", Content{
		Text: "평문 본문",
		HTML: "<p>HTML 본문</p>",
	})
	require.NoError(t, err)

	header, body := parse(t, raw)

	mediaType, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	require.NoError(t, err)
	assert.Equal(t, "multipart/alternative", mediaType)
	require.NotEmpty(t, params["boundary"])

	reader := multipart.NewReader(strings.NewReader(body), params["boundary"])

	type part struct {
		contentType string
		body        string
	}
	parts := []part{}
	for {
		p, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		buf, err := io.ReadAll(quotedprintable.NewReader(p))
		require.NoError(t, err)

		parts = append(parts, part{p.Header.Get("Content-Type"), string(buf)})
	}

	// RFC 2046 은 뒤에 오는 part 를 더 나은 표현으로 본다. HTML 이 뒤여야 한다
	require.Len(t, parts, 2)
	assert.Equal(t, part{"text/plain; charset=utf-8", "평문 본문"}, parts[0])
	assert.Equal(t, part{"text/html; charset=utf-8", "<p>HTML 본문</p>"}, parts[1])
}

func TestSendWithoutHost(t *testing.T) {
	sender := New(&Config{})

	err := sender.Send(context.Background(), "from@example.com", []string{"to@example.com"}, "제목", Content{
		Text: "본문",
	})
	assert.ErrorContains(t, err, "smtp host is not configured")
}

func TestSendRejectsBadInput(t *testing.T) {
	sender := New(&Config{Host: "smtp.example.com"})
	ctx := context.Background()

	t.Run("수신자 없음", func(t *testing.T) {
		err := sender.Send(ctx, "from@example.com", nil, "제목", Content{Text: "본문"})
		assert.ErrorContains(t, err, "no recipient")
	})
	t.Run("발신 주소가 주소가 아님", func(t *testing.T) {
		err := sender.Send(ctx, "not-an-address", []string{"to@example.com"}, "제목", Content{Text: "본문"})
		assert.ErrorContains(t, err, "not-an-address")
	})
	t.Run("수신 주소에 개행", func(t *testing.T) {
		err := sender.Send(ctx, "from@example.com",
			[]string{"to@example.com\r\nBcc: attacker@example.com"}, "제목", Content{Text: "본문"})
		assert.Error(t, err)
	})
}

// SMTP 서버에 붙기 전에 취소를 알아채는지 본다.
// 없는 호스트를 쓰므로, 취소를 안 보면 연결 에러가 대신 나온다
func TestSendCanceledContext(t *testing.T) {
	sender := New(&Config{Host: "smtp.invalid"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sender.Send(ctx, "from@example.com", []string{"to@example.com"}, "제목", Content{Text: "본문"})
	assert.ErrorIs(t, err, context.Canceled)
}

func TestAddrDefaultPort(t *testing.T) {
	assert.Equal(t, "smtp.example.com:25", New(&Config{Host: "smtp.example.com"}).addr())
	assert.Equal(t, "smtp.example.com:587", New(&Config{Host: "smtp.example.com", Port: 587}).addr())
}
