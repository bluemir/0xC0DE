# ADR-0002: Token 의 자격증명을 종류별 구조체로 나눠 한 컬럼에 담는다

- 상태: 채택
- 날짜: 2026-09-14
- 관련: [ADR-0001](ADR-0001-passkey-login.md) 의 "2. credential 은 기존 `Token` 테이블에 `kind = "passkey"` 로 저장한다" 를 대체한다

## 배경

ADR-0001 에서 passkey credential 을 `Token` 테이블에 얹으면서 `CredentialID`, `Credential`,
`Label` 세 컬럼을 더했다. 그 결과 한 행이 쓰는 필드가 `Kind` 에 따라 갈린다.

- `password`, `access-key` 행: `HashedSecret` 만 쓰고 passkey 컬럼 3개가 빈다
- `passkey` 행: passkey 컬럼 3개만 쓰고 `HashedSecret` 이 빈다

ADR-0001 에도 단점으로 적어 뒀지만, 인증 수단이 더 늘면 빈 컬럼이 계속 늘어난다.
Go 쪽에서도 `token.Label` 을 password 토큰에서 읽는 코드가 컴파일에 걸리지 않는다.

## 결정

### 1. 자격증명을 `Secret` 구조체로 옮기고 한 컬럼에 직렬화한다

```go
type Token struct {
	Username  string
	Kind      TokenKind
	Index     int
	ExpiredAt *time.Time
	CreatedAt time.Time

	CredentialID []byte `gorm:"uniqueIndex"`
	Secret       Secret `gorm:"type:bytes;serializer:gob"`
}

type Secret struct {
	Password  *PasswordSecret
	AccessKey *AccessKeySecret
	Passkey   *PasskeySecret
}
```

해당하는 것 하나만 non-nil 이고, 나머지는 nil 이다. nil 은 gob 왕복에서 그대로 보존된다.

- 장점: 빈 컬럼이 사라진다. 인증 수단을 더해도 컬럼과 마이그레이션이 늘지 않는다.
  종류별 필드에 접근하려면 `token.Secret.Passkey` 를 거쳐야 해서 잘못된 조합이 눈에 띈다.
- 단점: `Secret` 안의 값으로는 쿼리도 인덱스도 걸 수 없다. DB 에서 눈으로 읽을 수 없다.

### 2. `CredentialID` 는 `Secret` 밖에 남긴다

passkey 로그인은 username 을 받지 않고 authenticator 가 준 credential ID 로 사용자를 역추적한다.
`Secret` 안에 넣으면 unique index 를 걸 수 없어 이 조회가 불가능하다.
`Kind` 와 같은 성격 — 조회를 위한 키 — 으로 보고 컬럼으로 둔다.

passkey 가 아닌 행에서는 NULL 이다. 빈 컬럼이 완전히 없어지지는 않았다.

### 3. `Secret` 의 필드는 `TokenKind` 와 1:1 로 둔다

`password` 와 `access-key` 는 지금 둘 다 bcrypt 해시 하나가 전부라 한 구조체로 묶을 수도 있다.
그렇게 하면 `Token.Validate` 가 분기 없이 한 줄로 끝나지만, "공유 비밀" 같은 중간 개념이 하나 생기고
nil 조합과 `Kind` 가 1:1 로 떨어지지 않는다.

수단마다 따로 둔다. `Validate` 는 non-nil 인 쪽을 고르는 분기를 갖지만,
`Secret` 을 보면 어떤 수단인지가 그대로 드러나고 둘의 저장 형태가 갈라져도 영향이 없다.

### 4. 직렬화는 gob 을 쓴다

`User.Labels` 가 이미 `serializer:gob` 이다. 새 포맷을 섞지 않는다.
필드 추가는 안전하지만 타입 변경은 기존 행을 읽지 못하게 만든다.

### 5. `Kind` 와 `Secret` 의 정합성은 `createToken` 에서 본다

`Kind` 는 조회용 인덱스일 뿐이라 `Secret` 과 따로 놀 수 있다.
모든 토큰 생성이 `createToken` 하나를 지나므로 저장 직전에 확인한다.
`Secret` 에서 non-nil 인 것이 정확히 하나이고 그것이 `Kind` 와 같아야 하며,
`CredentialID` 는 `passkey` 일 때만 채워져 있어야 한다.

`TokenKind` 는 이로써 닫힌 집합이 된다. 새 인증 수단은 상수, `Secret` 필드,
`validateSecret` 의 case 세 곳을 같이 고쳐야 한다.

## GORM 사용상 제약

`serializer` 가 붙은 필드는 다음 두 경로에서 조용히 우회된다. 실측으로 확인했다.

- `Update("secret", value)` — serializer 를 타지 않고 구조체가 드라이버로 넘어가 실패한다.
  대신 `Updates(Token{Secret: ...})` 를 쓴다.
- `clause.Assignments(map[string]any{...})` — 값을 직접 넣으므로 마찬가지다.
  대신 `clause.AssignmentColumns([]string{"secret"})` 로 INSERT 에 쓰인 값을 재사용한다.

별개로 `db.Save(&token)` 은 쓸 수 없다. `Token` 의 복합 primary key 중 `Index` 가 0 이면
GORM 이 새 행으로 보고 INSERT 를 내보낸다.

## 결과

- 테이블은 `username, kind, index, expired_at, created_at, credential_id, secret` 이 된다.
- `Token.HashedSecret`, `Token.Credential`, `Token.Label` 은 사라졌다.
  각각 `Secret.Password.HashedSecret`(또는 `Secret.AccessKey.HashedSecret`),
  `Secret.Passkey.Credential`, `Secret.Passkey.Label` 이다.
- 기존 행을 옮기는 마이그레이션은 만들지 않았다. 운영 중인 DB 가 없다.
