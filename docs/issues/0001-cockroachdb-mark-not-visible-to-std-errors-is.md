# 0001: `errors.Mark` 는 표준 `errors.Is` 에 잡히지 않는다

- 대상: `github.com/cockroachdb/errors`
- 날짜: 2026-10-15
- 결론: 라이브러리 설계대로다. 버그가 아니므로 위에 보고하지 않는다

## 증상

`errors.Mark(cause, sentinel)` 로 표시한 에러를 표준 `errors.Is` 로 물으면 `false` 다.
cockroachdb 의 `errors.Is` 로 물어야 `true` 다.

이것 때문에 `internal/server/middleware/errs/code.go` 가 표준 `errors` 를 import 하던 동안
`auth` 가 `Mark` 로 표시해 올린 인증 실패가 401 이 아니라 **500** 으로 나갔다.
`internal/server/backend/auth/passkey.go` 의 두 곳(`FinishPasskeyRegistration`,
`FinishPasskeyLogin`)이 그 상태였고, webauthn 내부 에러 메시지가 응답 본문에 실렸다.

`testify` 의 `assert.ErrorIs` 도 표준 `errors.Is` 를 쓰므로 같은 이유로 통과하지 못한다.
그래서 `Mark` 로 표시한 것은 평범한 테스트로는 검증되지 않는다.

## 재현

```go
var sentinel = errors.New("unauthorized")

cause := errors.New("crypto/bcrypt: mismatch")

marked := errors.Mark(cause, sentinel)       // cockroachdb/errors
stderrors.Is(marked, sentinel)               // false
errors.Is(marked, sentinel)                  // true
assert.ErrorIs(t, marked, sentinel)          // 실패
```

비교 대상:

| 감싸는 방법 | 표준 `Is` | cockroachdb `Is` | testify | `Error()` |
|---|---|---|---|---|
| `Mark(cause, sentinel)` | false | true | 실패 | `crypto/bcrypt: mismatch` |
| `Join(sentinel, cause)` | true | true | 통과 | `unauthorized\ncrypto/bcrypt: mismatch` |
| `WithSecondaryError(sentinel, cause)` | true | true | 통과 | `unauthorized` |

## 우리 코드가 아니라는 근거

`Mark` 는 `markers.Mark` 로 가고, 그 표식은 `errors.Is` 의 cockroachdb 구현이
따로 들여다보는 자리에 붙는다. 표준 `errors.Is` 는 `Unwrap()` 사슬과
`Is(error) bool` 메서드만 보므로 그 자리를 알지 못한다.
표준 라이브러리에 그 표식을 읽을 길이 없다.

문서에도 마커는 자기네 `Is` 와 함께 쓰라고 되어 있다. 호환을 약속한 적이 없다.

## 곁다리

`errors.Join(sentinel, nil)` 은 **nil 이 아니다.** `sentinel` 하나를 담은 에러를 돌려준다.
`Join(nil, nil)` 만 nil 이다. 그래서

```go
return errors.Join(ErrUnauthorized, bcrypt.CompareHashAndPassword(...))
```

처럼 쓰면 비밀번호가 맞는 경우까지 에러가 되어 **모든 로그인이 실패한다.**
`errors.Mark(nil, sentinel)` 은 nil 을 그대로 돌려주므로 이 함정이 없다.
`Join` 으로 바꿀 때는 `err != nil` 안에서만 감싸야 한다.

## 우리가 한 일

- `code.go`, `render.go`, `handler/posts.go` 의 `"errors"` import 를
  `github.com/cockroachdb/errors` 로 바꿨다. 저장소 안에서는 마커가 읽힌다
- 그래도 `auth` 가 올리는 인증 실패는 `Mark` 가 아니라 `errors.Join` 으로 섞는다.
  저장소 밖(테스트의 `assert.ErrorIs` 를 포함해)은 여전히 표준 `Is` 를 쓰기 때문이다.
  [ADR-0009](../adr/ADR-0009-mark-auth-failures-with-join.md)
