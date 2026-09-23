// Package mail 은 SMTP 로 메일을 한 통 보낸다.
//
// 계정 복구 코드나 안내 메일처럼 "한 번에 한 통, 첨부 없는 본문"이 전부라서
// 라이브러리를 들이지 않고 net/smtp 와 mime/multipart 로 직접 만들었다.
// 어떤 요구사항이 생기면 라이브러리로 갈아탈지는 ADR-0005 에 적어두었다.
package mail

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
)

// defaultPort 는 Port 를 비워뒀을 때 쓰는 SMTP 포트다.
// 인증이 붙은 릴레이는 보통 587 이므로 그때는 설정에 명시해야 한다.
const defaultPort = 25

type Config struct {
	// 비면 Send 가 에러를 낸다. 메일을 조용히 버리지 않는다
	Host string
	Port int
	// From 은 이 서버가 보내는 메일의 발신 주소다.
	// Send 는 발신 주소를 인자로 받는다. 이 값은 호출측이 기본으로 쓰라고 두는 것이다.
	// 실존하지 않는 주소로 보내면 발송이 실패하고 발송 IP 의 평판이 깎이므로 기본값을 두지 않는다
	From string
	// 비면 인증 없이 보낸다. IP ACL 로만 통제하는 사내 릴레이가 그렇다
	Username string
	Password string
}

// Content 는 본문이다. HTML 이 비면 text/plain 한 벌로,
// 차 있으면 multipart/alternative 로 둘 다 실어 보낸다.
type Content struct {
	Text string
	HTML string
}

type Sender struct {
	conf Config
}

func New(conf *Config) *Sender {
	return &Sender{conf: *conf}
}

// Send 는 메일 한 통을 보낸다.
//
// from 과 to 는 "a@example.com" 과 "이름 <a@example.com>" 둘 다 받는다.
// net/smtp 에는 context 를 받는 API 가 없어서 ctx 는 보내기 전에 한 번만 본다.
// 발송이 시작된 뒤에는 취소되지 않는다.
func (s *Sender) Send(ctx context.Context, from string, to []string, title string, content Content) error {
	if s.conf.Host == "" {
		return errors.New("smtp host is not configured")
	}
	if len(to) == 0 {
		return errors.New("no recipient")
	}
	if err := ctx.Err(); err != nil {
		return errors.WithStack(err)
	}

	// 봉투(envelope)에는 표시 이름을 뺀 주소만 들어간다.
	// ParseAddress 는 주소에 CR/LF 가 섞인 헤더 인젝션도 함께 걸러낸다.
	fromAddress, err := mail.ParseAddress(from)
	if err != nil {
		return errors.Wrapf(err, "sender %q", from)
	}
	recipients := make([]string, 0, len(to))
	for _, addr := range to {
		parsed, err := mail.ParseAddress(addr)
		if err != nil {
			return errors.Wrapf(err, "recipient %q", addr)
		}
		recipients = append(recipients, parsed.Address)
	}

	msg, err := build(from, to, title, content)
	if err != nil {
		return err
	}

	var auth smtp.Auth
	if s.conf.Username != "" {
		auth = smtp.PlainAuth("", s.conf.Username, s.conf.Password, s.conf.Host)
	}

	// 서버가 STARTTLS 를 광고하면 net/smtp 가 알아서 쓴다.
	// 465(implicit TLS)는 net/smtp 로 못 붙는다
	if err := smtp.SendMail(s.addr(), auth, fromAddress.Address, recipients, msg); err != nil {
		return errors.Wrapf(err, "send mail to %v via %s", recipients, s.addr())
	}

	return nil
}

// From 은 설정에 적힌 발신 주소다. 호출측이 Send 의 from 으로 그대로 넘긴다.
func (s *Sender) From() string {
	return s.conf.From
}

func (s *Sender) addr() string {
	port := s.conf.Port
	if port == 0 {
		port = defaultPort
	}
	return fmt.Sprintf("%s:%d", s.conf.Host, port)
}

// build 는 RFC 5322 메시지를 만든다.
// 제목은 RFC 2047 로 인코딩해 한글이 깨지지 않게 한다.
func build(from string, to []string, title string, content Content) ([]byte, error) {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, "From: %s\r\n", from)
	fmt.Fprintf(buf, "To: %s\r\n", strings.Join(to, ", "))
	fmt.Fprintf(buf, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", title))
	fmt.Fprintf(buf, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(buf, "MIME-Version: 1.0\r\n")

	if content.HTML == "" {
		fmt.Fprintf(buf, "Content-Type: text/plain; charset=utf-8\r\n")
		fmt.Fprintf(buf, "Content-Transfer-Encoding: quoted-printable\r\n")
		fmt.Fprintf(buf, "\r\n")

		if err := writeBody(buf, content.Text); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	parts := multipart.NewWriter(buf)

	fmt.Fprintf(buf, "Content-Type: multipart/alternative; boundary=%s\r\n", parts.Boundary())
	fmt.Fprintf(buf, "\r\n")

	// RFC 2046 은 뒤에 오는 part 를 더 나은 표현으로 본다. HTML 을 뒤에 둔다
	for _, part := range []struct {
		contentType string
		body        string
	}{
		{"text/plain; charset=utf-8", content.Text},
		{"text/html; charset=utf-8", content.HTML},
	} {
		w, err := parts.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {part.contentType},
			"Content-Transfer-Encoding": {"quoted-printable"},
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if err := writeBody(w, part.body); err != nil {
			return nil, err
		}
	}

	if err := parts.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	return buf.Bytes(), nil
}

// writeBody 는 본문을 quoted-printable 로 쓴다.
// 8bit 로 그냥 실으면 한글 본문이 SMTP 의 998 바이트 줄 제한에 걸릴 수 있다.
func writeBody(w io.Writer, body string) error {
	qp := quotedprintable.NewWriter(w)
	if _, err := qp.Write([]byte(body)); err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(qp.Close())
}
