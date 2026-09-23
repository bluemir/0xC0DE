# ADR-0005: 메일은 net/smtp 로 직접 보내고 라이브러리를 들이지 않는다

- 상태: 채택
- 날짜: 2026-09-22

## 배경

계정 복구와 안내 메일을 보낼 자리가 필요하다. 첫 사용처는 비밀번호 재설정 링크 발송이고,
passkey 만으로 가입한 계정이 마지막 passkey 를 잃었을 때의 복구 경로도 여기에 기댄다.

보낼 메일의 모양이 정해져 있다.

- 한 번에 한 통, 수신자는 사실상 한 명
- 첨부 없음, 인라인 이미지 없음
- 본문은 짧은 안내문과 링크 하나

## 결정

**`internal/mail` 에 의존성 없이 `net/smtp` + `mime/multipart` 로 직접 만든다.**

```go
type Content struct {
	Text string
	HTML string
}

func (s *Sender) Send(ctx context.Context, from string, to []string, title string, content Content) error
```

구현은 173줄, 테스트는 184줄이다.

### 1. 왜 라이브러리가 아닌가

`gopkg.in/gomail.v2` 와 `wneessen/go-mail` 을 봤다. 둘이 대신 해주는 일은
MIME 조립, 헤더 인코딩, TLS 협상, 첨부, 재시도다. 이 중 지금 필요한 것은 앞의 셋뿐이고,
표준 라이브러리가 그 셋을 각각 `mime/multipart`, `mime.QEncoding`, `smtp.SendMail`
(STARTTLS 자동)로 이미 덮는다. 남는 것은 그걸 순서대로 부르는 코드다.

같은 판단을 다른 프로젝트(logbook)에서 한 번 했고 그쪽은 `text/plain` 한 벌이라 더 짧다.
여기는 HTML 본문을 같이 보내기로 해서 `multipart/alternative` 조립이 더 붙었다.

`gomail.v2` 는 버전이 `v2.0.0-20160411...` 에서 멈춰 있어 새로 들일 후보가 아니다.
라이브러리로 간다면 `wneessen/go-mail` 이다.

### 2. 직접 쓰면서 조심한 것

- **제목 인코딩** — `mime.QEncoding.Encode("utf-8", title)`. 안 하면 한글 제목이 깨진다.
  덤으로 헤더 인젝션도 막힌다. `QEncoding` 은 CR/LF 를 인코딩 대상으로 보기 때문에
  제목에 `\r\nBcc:` 를 넣어도 헤더가 하나 더 생기지 않는다.
- **주소 검증** — `net/mail.ParseAddress`. `From`/`To` 로 들어오는 인젝션을 여기서 막는다.
  봉투(envelope)에는 표시 이름을 뺀 주소만 넘긴다.
- **본문 인코딩** — `quoted-printable`. 8bit 로 실으면 한글 본문이 SMTP 의 998 바이트
  줄 제한에 걸릴 수 있다.
- **part 순서** — RFC 2046 은 뒤에 오는 part 를 더 나은 표현으로 본다. `text/plain` 을 앞,
  `text/html` 을 뒤에 둔다.

### 3. 본문은 multipart/alternative 로 두 벌을 보낸다

`Content{Text, HTML}` 을 받아 `text/plain` 과 `text/html` 을 같이 싣는다.

`alternative` 는 "같은 내용의 다른 표현이니 하나만 골라 보여라" 다(`mixed` 가 "다 보여라" 다).
수신자 화면에 두 번 나오지 않는다. RFC 2046 이 충실도가 낮은 것부터 배치하라고 해서
평문을 앞, HTML 을 뒤에 둔다.

**지금 보내는 메일에는 HTML 이 하는 일이 없다.** 복구 메일은 안내문 몇 줄과 링크 하나이고
`recovery.html` 은 그것을 `<p>` 로 감싼 것뿐이다. 평문 한 벌로 줄이면 `Content` 구조체,
multipart 조립, `html/template` 파서와 이스케이프 테스트까지 120줄 남짓이 사라진다.
`Send` 의 마지막 인자도 그냥 문자열이 된다.

그런데도 두 벌을 유지하는 이유는 하나다. **앞으로 서식 있는 메일을 보낼 생각이라서** 다.
나중에 다시 넣는 것보다 지금 자리를 두는 편이 싸다고 봤다.

평문 한 벌만 보내는 것도 정상이다(Gmail 의 "일반 텍스트 모드" 가 그렇다). 반대로 평문 대체본
없는 HTML-only 는 권하지 않는다. `MIME_HTML_ONLY` 같은 스팸 규칙에 걸리고 평문으로 보는
클라이언트에서 깨진다. 가중치가 크지는 않지만 얻는 것도 없다.

**다시 볼 조건** — 한동안 링크 한 줄짜리 메일만 보내고 있으면 평문 한 벌로 줄인다.
그때 지워야 할 것은 위에 적은 그 목록이다.

### 4. 받아들인 한계

| 한계 | 이유 |
|------|------|
| `ctx` 는 보내기 전에 한 번만 본다 | `net/smtp` 에 context 를 받는 API 가 없다. 발송이 시작된 뒤에는 취소되지 않는다 |
| 465(implicit TLS) 로 못 붙는다 | `smtp.SendMail` 이 평문으로 연결한 뒤 STARTTLS 를 협상한다. 587 이나 25 를 쓴다 |
| 첨부·인라인 이미지 없음 | `multipart/mixed`, `multipart/related` 를 안 만든다 |
| 재시도·큐잉 없음 | `Send` 가 실패하면 에러를 돌려주고 끝이다. 호출측이 정한다 |

`Host` 가 비면 `Send` 가 에러를 낸다. 메일을 조용히 버리지 않는다.
설정을 빼먹은 개발 환경에서 복구 메일을 부르면 그 자리에서 드러난다.

### 5. 라이브러리로 갈아탈 조건

아래 중 하나라도 생기면 다시 본다. 그때는 `wneessen/go-mail` 부터 본다.

- **첨부파일이나 인라인 이미지** — `multipart/mixed`·`multipart/related` 중첩을 직접 쌓게 된다
- **DKIM 서명** — 직접 구현할 것이 아니다
- **465(implicit TLS) 나 OAuth2 SMTP 인증** — Gmail·Office365 를 직접 릴레이로 쓰게 될 때다.
  `smtp.PlainAuth` 로는 안 된다
- **대량 발송** — 연결 재사용, 통당 delay, 재시도가 필요해지는 순간.
  지금은 통마다 새로 연결한다
- **발송 큐** — 실패한 메일을 나중에 다시 보내야 한다면 `jobs` 와 엮이면서 구조가 달라진다

갈아탈 때 바뀌는 것은 `internal/mail` 안쪽뿐이다. `Send` 의 시그니처와 `Content` 는
그대로 둘 수 있게 짰다. 호출측은 `mail.Content{Text:..., HTML:...}` 만 안다.

## 함께 정한 것

- **위치는 `internal/mail`** — DB 도 pubsub 도 안 쓰므로 `internal/server/backend` 밖에 둔다.
  설정은 `backend.Config` 가 아니라 `server.Config` 에 직접 붙는다.
- **인증은 `Config` 에만 둔다** — `Send` 에 옵션으로 넘기지 않는다.
  자격증명은 호출마다 달라지는 값이 아니다. `Username` 이 비면 인증 없이 보낸다
- **수신자는 `[]string`** — 지금은 한 명뿐이지만 나중에 시그니처를 바꾸지 않으려고
