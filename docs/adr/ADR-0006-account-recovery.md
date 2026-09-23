# ADR-0006: 계정 복구는 메일 링크로 열리는 "복구 세션" 으로 한다

- 상태: 채택
- 날짜: 2026-09-22
- 관련: [ADR-0002](ADR-0002-token-secret-per-kind.md) 의 자격증명 모델, [ADR-0005](ADR-0005-smtp-without-library.md) 의 메일 발송

## 배경

로그인 수단을 잃은 사용자를 되살릴 길이 없었다.

- 비밀번호를 잊으면 방법이 없다. 비밀번호 변경 API 자체가 없었다
- passkey 만으로 가입한 계정이 마지막 기기를 잃으면 영영 못 들어온다
  (`docs/tasks.md` 의 "마지막 로그인 수단을 지우지 못하게 막기" 가 가리키던 구멍이다)

사용자에게 연락할 수단도 없었다. `auth.User` 는 `Name`, `Salt`, `Handle`, `Groups`, `Labels`
뿐이고 가입 폼은 username/password 만 받았다.

## 결정

### 1. `User.Email` 을 더한다. 선택 입력이고 확인하지 않는다

```go
Email string `gorm:"index;size:256" json:"email" expr:"email"`
```

- **선택 입력** — 비워도 가입된다. passkey 가입(`FinishPasskeyRegistration`)은 ceremony 중에
  이메일을 받을 자리가 없고, 이미 있는 계정도 빈 값으로 남는다. 필수로 만들면 이 둘을 다
  건드려야 한다. 대신 이메일 없는 계정은 복구할 수 없고, 그 사실을 화면에 적었다.
- **uniqueIndex 가 아니다** — 한 사람이 여러 계정을 가질 수 있다. 같은 주소가 여럿이면
  `FindUserByEmail` 은 먼저 만들어진 것을 돌려준다.
- **확인 절차 없음** — 적은 값을 그대로 믿는다. 오타를 내면 복구 메일이 엉뚱한 곳으로 가고,
  남의 주소를 적을 수도 있다. 확인 메일은 후속 과제로 남겼다.
- 등록 경로는 둘이다. 가입 폼과 `PATCH /api/v1/users/me`(`/users/settings`).
  둘째가 없으면 이미 가입한 계정이 영원히 복구 불가라 같이 넣었다.

### 2. 복구 비밀은 `Token` 이 아니라 별도 테이블에 둔다

```go
type Recovery struct {
	Username     string `gorm:"primaryKey;size:256"`
	HashedSecret []byte
	ExpiredAt    time.Time
	CreatedAt    time.Time
}
```

처음에는 `TokenKindRecovery` 를 더해 `Token` 에 넣으려 했다. `ExpiredAt`, bcrypt 해싱,
`Validate`, `RevokeToken` 이 이미 있어서 재사용이 쉬워 보였다. 다음 이유로 접었다.

- `ListToken(username)` 에 딸려 나온다. 복구 통행증이 로그인 자격증명 목록에 섞인다
- `createToken` 은 `(username, kind)` 마다 `index` 를 늘린다. 복구를 요청할 때마다 행이 쌓이고
  만료된 것을 치우는 청소가 따로 필요해진다
- `Secret` 은 "로그인에 쓰는 자격증명" 이다. 복구 비밀은 로그인 수단이 아니라 수단을 다시
  세우기 위한 임시 통행증이다

**사용자당 한 행**이라 재발급이 덮어쓰기가 된다. 행이 쌓이지 않고, 재발송 간격도 `CreatedAt`
하나로 잰다.

비밀은 `crypto/rand` 32 byte 를 base64url 로 적은 것이다. `util.RandomString` 은 `math/rand` 라
예측할 수 있어 쓰지 않았다.

### 3. 링크를 열면 "복구 세션" 이 열린다. 로그인이 아니다

```
session[recovery] = {Username}
```

링크를 로그인으로 바꾸는 쪽이 코드가 훨씬 적었다. 기존 화면을 그대로 쓰면 되기 때문이다.
그러지 않은 이유는 권한의 크기다. 메일함을 잠깐 들여다본 사람이 곧바로 그 계정의 모든 것을
할 수 있게 된다. 복구 세션에서 할 수 있는 일은 인증 수단을 세우는 것뿐이다.

- `POST /api/v1/recover/password` — 비밀번호 설정
- `POST /api/v1/passkeys/register/*` — passkey 등록.
  `PasskeyRegisterBegin` 이 복구 세션도 계정 소유 증명으로 인정하도록 고쳤다.
  원래는 이미 있는 계정에 passkey 를 붙이려면 로그인해야 했다

둘 중 하나를 마치면 복구 비밀을 폐기하고 그때 로그인시킨다. 링크는 한 번만 쓰인다.

**passkey 는 지우지 않는다.** 계정 탈취를 가정하면 전부 지우는 쪽이 안전하지만, 그러면
메일함만 뚫린 공격자가 정상 사용자의 passkey 를 지울 수 있다. access-key 도 지금은 두는데,
이건 다시 볼 값이 있다(아래).

### 4. 계정이 있는지 응답으로 드러내지 않는다

없는 이메일로 요청해도, 최근에 이미 보냈어도 같은 답을 준다.

발송은 `jobs.Manager` 에 넘겨 백그라운드로 보낸다. 응답을 기다리는 동안 SMTP 와 bcrypt 를
거치면 **계정이 있는 요청만 느려져 타이밍으로 존재 여부가 샌다.** 덤으로 job 행이 남아
언제 누구에게 보냈는지 추적할 수 있다.

### 5. 15분 만료, 15분에 한 통

`auth.RecoveryLifetime`, `auth.RecoveryResendInterval` 이다. 메일을 바로 확인하는 사람을
기준으로 잡았다. 재발송 간격은 메일 폭탄을 막는다.

`IssueRecovery` 는 `SELECT ... FOR UPDATE` 로 직전 발급을 잠그고 본다. 동시에 두 번 눌러
둘 다 통과하는 것을 막는다.

### 6. 메일 문구는 템플릿 파일에 둔다

`internal/server/handler/mail-templates/` 에 메일 한 통당 세 파일을 두고 `//go:embed` 한다.

```
recovery.title   제목 한 줄
recovery.txt     평문 본문
recovery.html    HTML 본문
```

웹 페이지 템플릿(`assets/html-templates`)과 섞지 않았다. 메일은 웹 페이지와 배포 단위도
렌더 방식도 다르고, 보내는 코드 바로 옆에 있는 편이 찾기 쉽다.

**평문과 HTML 을 다른 파서로 읽는다.** 평문은 `text/template`, HTML 은 `html/template` 이다.
한 파서로 둘 다 처리하면 평문 본문에 `&amp;` 가 튀어나오거나, 반대로 HTML 본문에 사용자가
적은 `<script>` 가 그대로 들어간다. username 은 사용자가 정하는 값이라 뒤쪽이 실제 위험이다.
옮기기 전에는 Go 문자열에 `fmt.Sprintf` 로 끼워 넣고 있어서 이 이스케이프가 없었다.

## 받아들인 한계

| 한계 | 비고 |
|------|------|
| 이메일을 확인하지 않는다 | 오타면 메일이 엉뚱한 곳으로 간다. 남의 주소도 적을 수 있다 |
| 이메일 없는 계정은 복구 불가 | 화면에 적었다. passkey 가입자가 특히 그렇다 |
| 복구 요청 자체에 rate limit 이 없다 | 계정당 15분 간격만 있다. 모르는 주소로 퍼붓는 것은 안 막는다 |
| 복구 링크가 URL 경로에 있다 | 브라우저 기록과 Referer 에 남는다. 15분이면 만료된다 |
| access-key 를 지우지 않는다 | 탈취당한 계정이면 공격자가 발급해 둔 키가 살아남는다 |

## 다시 볼 조건

- **가입에 이메일을 필수로 만들 때** — 확인 메일이 같이 와야 한다. 그때 `EmailVerifiedAt` 을
  더하고, 확인된 주소만 복구에 쓴다
- **탈취 사례가 나올 때** — 복구 완료 시 access-key 를 전부 폐기할지 다시 본다.
  지금 안 하는 것은 정상 사용자의 자동화가 조용히 끊기는 쪽이 더 흔할 것으로 봐서다
- **복구 요청이 남용될 때** — IP 기준 rate limit 이 필요하다. 지금은 계정당 간격뿐이다
