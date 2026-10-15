# ADR-0009: 인증 실패는 `errors.Join` 으로 `ErrUnauthorized` 를 섞어 올린다

- 상태: 채택
- 날짜: 2026-10-15
- 관련: [issue 0001](../issues/0001-cockroachdb-mark-not-visible-to-std-errors-is.md)

## 배경

`Manager.Default` 로 잘못된 비밀번호를 넣으면 이렇게 나갔다.

```
500 Internal Server Error
{"Error":"crypto/bcrypt: hashedPassword is not the hash of the given password"}
```

`Token.Validate` 가 `bcrypt.CompareHashAndPassword` 의 에러를 그대로 올리고,
`errs` 미들웨어의 `code()` 가 아는 에러가 아니라 500 으로 떨어뜨린 것이다.
상태 코드가 틀렸고, 내부 에러 메시지가 응답에 실렸다.

고치려고 보니 `auth` 에는 이미 `errors.Mark(err, ErrUnauthorized)` 라는 선례가
`passkey.go` 두 곳에 있었다. **그런데 그 두 곳도 똑같이 500 으로 나가고 있었다.**
`code.go` 가 표준 `errors` 를 import 하는데 cockroachdb 의 마커는 그쪽 `Is` 를
거쳐야만 보이기 때문이다. 자세한 것은 [issue 0001](../issues/0001-cockroachdb-mark-not-visible-to-std-errors-is.md) 에 적었다.

## 결정

### 1. 매핑하는 쪽은 cockroachdb 의 `errors` 를 쓴다

`code.go`, `render.go`, `handler/posts.go` 의 `"errors"` import 를 바꿨다.
저장소 전체가 `cockroachdb/errors` 로 감싸 올리는데 판정하는 자리만 표준을 쓰면,
표시해 둔 것이 묻힌다.

바꾸기 전에 `code()` 의 동작을 17가지 경우로 고정해 두고(`errs/code_test.go`)
바꾼 뒤에도 같은지 확인했다. 그 패키지에는 테스트가 없었다.

### 2. 그래도 인증 실패는 `Mark` 가 아니라 `errors.Join` 으로 섞는다

```go
if err := bcrypt.CompareHashAndPassword(hashedSecret, []byte(unhashedSecret)); err != nil {
	return errors.Join(ErrUnauthorized, err)
}
return nil
```

1 로 저장소 안은 마커가 읽히게 됐지만, **저장소 밖은 여전히 표준 `errors.Is` 를 쓴다.**
`testify` 의 `assert.ErrorIs` 가 그렇다. `Mark` 로 표시한 것은 평범한 테스트로
검증할 수 없고, 검증되지 않는 표식은 조용히 썩는다. 실제로 `passkey.go` 의 두 곳이
그렇게 썩어 있었다.

`Join` 은 표준과 cockroachdb 양쪽에서 읽힌다. 내장 함수라 다음 사람이 덜 놀란다.

**버린 선택지: `errors.WithSecondaryError(ErrUnauthorized, cause)`.**
읽히는 것은 `Join` 과 같고, `Error()` 가 `"unauthorized"` 하나라서 bcrypt 메시지가
응답에 실리지 않는다는 점이 더 낫다. 고르지 않았다. 덜 알려진 API 라
다음에 읽는 사람이 왜 이것인지 다시 찾아야 한다. 메시지 노출은 `errs` 가
모든 에러의 `err.Error()` 를 응답에 담는 더 큰 문제의 일부이고,
인증 경로에서만 막아 봐야 반쪽이다. 그쪽을 통째로 보는 일은 따로 남겼다.

### 3. 적용 범위는 `Token.Validate` 의 두 실패와 `passkey.go` 의 두 곳이다

`Validate` 는 만료와 불일치를 모두 섞어 올린다. 부르는 쪽에는 둘 다
"이 비밀로는 통과 못 한다" 하나다. `// TODO` 가 붙어 있던 `"token is expired"` 도
여기서 같이 답했다.

`passkey.go` 의 `Mark` 두 곳은 `Join` 으로 맞췄다. `auth` 가 올리는 인증 실패가
한 모양이 된다.

## 함정

`errors.Join(sentinel, nil)` 은 **nil 이 아니다.** `sentinel` 하나를 담은 에러가 된다.
`Mark(nil, sentinel)` 이 nil 을 돌려주는 것과 다르다. 그래서

```go
return errors.Join(ErrUnauthorized, bcrypt.CompareHashAndPassword(...))
```

로 쓰면 비밀번호가 맞아도 실패가 된다. `err != nil` 안에서만 감싼다.
`Validate` 에 주석으로 남겼다.

## 결과

- 잘못된 비밀번호는 `401` 이다. passkey 등록·로그인 실패도 `401` 이다
- 네 자리 모두 뮤테이션으로 확인했다. 감싸는 것을 지우면 테스트가 깨진다
- **응답 본문에는 여전히 원인 메시지가 실린다.** `{"Error":"unauthorized\ncrypto/bcrypt: ..."}`
  가 나간다. `errs` 가 모든 에러의 `err.Error()` 를 담는 문제라 따로 본다.
  `docs/tasks.md` 에 남겼다
- `errs` 패키지에 `code()` 테스트가 생겼다. 그 전에는 없었다
