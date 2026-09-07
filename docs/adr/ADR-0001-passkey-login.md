# ADR-0001: Passkey(WebAuthn) 로그인 추가

- 상태: 채택
- 날짜: 2026-09-07

## 배경

기존 로그인 수단은 ID/PW 하나뿐이다.
`auth.Manager` 는 `Token` 을 `kind`(`password`, `access-key`)로 구분해 bcrypt 해시를 저장하고,
`handler.Login` 이 세션에 `*auth.User` 를 담는 구조다.

여기에 passkey 로그인을 더하면서 결정해야 할 것들이 있었다.

## 결정

### 1. WebAuthn 검증은 `github.com/go-webauthn/webauthn` 을 쓴다

attestation 파싱, CBOR/COSE 해석, origin·challenge·서명·sign counter 검증을 직접 구현하면
코드량이 크고 검증 로직 실수가 곧 인증 우회로 이어진다.

- 장점: 규격 대응과 보안 검증을 라이브러리가 담당한다.
- 단점: 의존성이 늘어난다(`go-webauthn/webauthn`, `go-webauthn/x`, `fxamacker/cbor`, `google/go-tpm`).

### 2. credential 은 기존 `Token` 테이블에 `kind = "passkey"` 로 저장한다

테이블을 새로 만들지 않고 인증 수단을 한 곳에서 관리한다.
`Token` 에 passkey 전용 컬럼을 더했다.

- `CredentialID` - 로그인 시 사용자를 찾는 키. unique index.
- `Credential` - `webauthn.Credential` 을 JSON 으로 저장. 라이브러리가 권장하는 두 방식 중
  "불투명한 직렬화 값" 쪽이다. 필드별 컬럼으로 펼치는 방식이 권장되지만, `Token` 을 재사용하는 이상
  nullable 컬럼이 열 개 넘게 늘어나므로 조회 키만 컬럼으로 두고 나머지는 묶었다.
- `Label` - 사용자가 기기를 구분하는 이름.

- 장점: 인증 수단 목록·폐기 흐름(`ListToken`, `RevokeToken`)을 그대로 쓴다.
- 단점: `Token` 의 의미가 "공유 비밀의 해시"에서 넓어졌고, passkey 가 아닌 행에는 비는 컬럼이 생긴다.
  공개키는 비밀이 아니므로 `HashedSecret` 에 담지 않고 별도 컬럼으로 뒀다.

### 3. username 없이 로그인한다 (discoverable credential)

등록할 때 `residentKey=required` 를 요구해 authenticator 가 credential 을 직접 들고 있게 하고,
로그인은 `BeginDiscoverableLogin` / `FinishPasskeyLogin` 으로 처리한다.
서버는 응답의 credential ID 로 사용자를 찾는다.

- 장점: 로그인 화면에서 username 입력 단계가 사라진다.
- 단점: authenticator 가 discoverable credential 을 지원해야 한다.

### 4. WebAuthn user handle 은 `User.Handle` 로 따로 둔다

규격은 handle 을 사용자에게 보이지 않는 불투명한 값으로 두라고 권한다.
username 을 그대로 쓰면 authenticator 에 사용자 이름이 남고 64 byte 제한도 걸린다.
`User` 에 32 byte 난수 `Handle` 을 추가하고 unique index 를 걸었다.
ID/PW 로 먼저 가입해 `Handle` 이 없는 사용자는 passkey 등록 시점에 채운다.

### 5. RP ID 와 origin 은 config 에 두고, 없으면 요청에서 추론한다

credential 은 RP ID 에 묶이므로 운영 환경에서는 `auth.passkey.id` 를 고정해야 한다.
설정이 비면 `location.Get(c)` 가 준 host/origin 을 쓴다. 로컬 개발에서 설정 없이 바로 쓰기 위한 것이다.

- 단점: 설정을 비운 채로 배포하면 reverse proxy 가 준 Host 헤더를 그대로 신뢰한다.
  운영 설정에서는 반드시 채운다.

### 6. ceremony 중간 상태는 세션 쿠키에 담는다

`BeginRegistration` / `BeginDiscoverableLogin` 이 준 `webauthn.SessionData` 는
JSON 으로 직렬화해 `gin-contrib/sessions` 세션에 넣고 finish 단계에서 꺼내 쓴다.
꺼낸 뒤에는 바로 지워 한 번만 쓰이게 한다.

challenge 는 비밀이 아니지만 클라이언트가 고칠 수 없어야 한다.
현재 세션 저장소는 `cookie.NewStore` 에 키 하나만 주므로 서명만 하고 암호화하지는 않는다.
무결성은 보장되므로 ceremony 용도로는 충분하다.

## 결과

- 기존 계정에 passkey 를 추가로 등록할 수 있다. ID/PW 로그인은 그대로 동작한다.
- passkey 만으로 가입할 수도 있다. 이 계정에는 `password` 토큰이 없어 `Manager.Default` 로는 로그인되지 않는다.
- 이미 있는 계정에 passkey 를 붙이려면 그 계정으로 로그인해야 한다.
  로그인 없이 임의의 username 으로 등록을 시작하면 남의 계정에 passkey 를 심을 수 있으므로,
  `PasskeyRegisterBegin` 은 이미 있는 username 에 대해 401 을 돌려준다.
