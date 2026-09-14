# ADR-0003: 소셜 로그인 자리를 Token 에 미리 만든다

- 상태: 채택
- 날짜: 2026-09-14
- 관련: [ADR-0002](ADR-0002-token-secret-per-kind.md) 의 `CredentialID` 를 `ExternalID` 로 일반화한다

## 배경

로그인 수단은 ID/PW 와 passkey 두 가지다. `Token.Kind` 주석에는 처음부터
`google, github...` 가 적혀 있었지만 자리는 없었다.

provider 와 주고받는 부분(authorization code 교환, 프로필 조회)은 아직 필요하지 않다.
다만 데이터 모델은 나중에 바꾸기 비싸므로 먼저 정해 둔다.
이 ADR 은 "우리가 OAuth 클라이언트가 되어 외부 계정으로 로그인받는" 경우만 다룬다.
우리가 OAuth 제공자가 되는 것과 외부 API 호출용 토큰을 보관하는 것은 범위 밖이다.

## 결정

### 1. provider 마다 `TokenKind` 를 하나씩 쓴다

```go
const (
	TokenKindGoogle TokenKind = "google"
	TokenKindGitHub TokenKind = "github"
)
```

`oauth` 하나를 두고 provider 이름을 `Secret` 안에 넣는 방법도 있다.
그러면 provider 를 더할 때 `Kind` 를 건드리지 않아도 되지만, 특정 provider 연결을 찾으려면
목록을 받아 `Secret` 을 훑어야 한다. `Kind` 는 primary key 의 일부이자 조회 키라는 성격에 맞지 않는다.

provider 마다 `Kind` 를 쓰면 `GetToken(user, TokenKindGoogle, 0)` 이 그대로 먹고,
`index` 채번도 provider 별로 따로 돈다.

- 단점: provider 를 더할 때마다 상수, `validateSecret`, `IsOAuth`, `ListOAuth` 네 곳에
  provider 이름이 나열된다. 한 곳으로 모을 수도 있지만 각각이 하는 일이 달라 그대로 뒀다.
  `TokenKind` 가 닫힌 집합이라 빠뜨리면 `createToken` 에서 막힌다.

### 2. `Secret.OAuth` 는 provider 가 공유한다

`Kind` 가 이미 어느 provider 인지 말해주므로 `OAuthSecret` 을 provider 마다 나누지 않는다.
ADR-0002 에서 `password` 와 `access-key` 를 나눈 것과는 반대인데, 저 둘은 같은 개념의
다른 수단이고 provider 들은 같은 수단의 다른 상대이기 때문이다.

이로써 `Secret` 필드와 `Kind` 의 대응이 1:1 이 아니게 되어, `validateSecret` 은
`Kind` 로 분기해 "그 `Kind` 에 맞는 것 하나만 채워졌는가" 를 본다.

### 3. `CredentialID` 를 `ExternalID` 로 일반화하고 `Kind` 와 묶어 인덱스를 건다

passkey 는 credential ID 로, 소셜 로그인은 provider 가 준 사용자 ID(subject)로
사용자를 역추적한다. 둘 다 "바깥에서 온 식별자로 우리 사용자를 찾는" 같은 일이다.

```go
Kind       TokenKind `gorm:"primaryKey;size:256;uniqueIndex:idx_tokens_external,priority:1"`
ExternalID []byte    `gorm:"uniqueIndex:idx_tokens_external,priority:2"`
```

인덱스는 `(kind, external_id)` 복합으로 건다. subject 는 provider 안에서만 유일해서
Google 의 `12345` 와 GitHub 의 `12345` 가 단일 인덱스에서는 충돌한다.
passkey 조회도 원래 `kind = ? AND credential_id = ?` 였으므로 쿼리 모양은 그대로다.

같은 외부 계정이 두 사용자에 붙는 것은 이 인덱스가 막는다.

- 단점: 컬럼을 새로 두는 대신 이름을 바꿨으므로 passkey 쪽 코드가 같이 바뀌었다.
- `password` 와 `access-key` 행에서는 여전히 NULL 이다.
  복합 인덱스에서 NULL 은 서로 다른 값으로 취급되어 여러 행이 공존한다.

### 4. `OAuthSecret` 에는 보여주기 위한 것만 담는다

```go
type OAuthSecret struct {
	Email       string
	Name        string
	RefreshedAt time.Time
}
```

로그인에 필요한 값은 `Kind` 와 `ExternalID` 가 다 가지고 있다.
여기 있는 것은 "어느 계정과 연결했는지" 를 목록에서 보여주기 위한 사본이다.
provider 쪽에서 바뀌면 오래된 값이 남으므로 `RefreshedAt` 으로 언제 받은 것인지 남긴다.

access token 과 refresh token 은 담지 않는다. 로그인만 할 것이라 필요 없고,
보관하면 유출 시 피해 범위가 우리 서비스 밖으로 넘어간다.
외부 API 를 대신 호출할 일이 생기면 그때 따로 결정한다.

### 5. 흐름은 만들지 않는다

`golang.org/x/oauth2` 의존성, state/PKCE, 토큰 교환, provider 설정은 넣지 않았다.
`Manager` 에는 연결을 만들고 찾고 끊는 것만 있다.

- `LinkOAuth` - 외부 계정을 사용자에 연결한다
- `FindOAuthUser` - provider 가 준 ID 로 사용자를 찾는다
- `UpdateOAuthProfile` - 프로필 사본을 다시 저장한다
- `ListOAuth` / `RevokeOAuth` - 연결 목록과 해제

가입 여부(연결된 사용자가 없을 때 계정을 만들 것인가)는 정하지 않았다.
passkey 는 `FinishPasskeyRegistration` 이 정하는데, 그에 해당하는 자리가 아직 없다.

## 결과

- `Token.CredentialID` 는 `Token.ExternalID` 가 되었다. unique index 는 `(kind, external_id)` 다.
- `TokenKind` 에 `google`, `github` 이 생겼다. 선언되지 않은 provider 는 저장되지 않는다.
- 새 의존성은 없다.
- 실제 로그인 흐름, provider 설정, 가입 정책은 남아 있다. `docs/tasks.md` 에 적었다.
