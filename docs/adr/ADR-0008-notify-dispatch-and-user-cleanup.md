# ADR-0008: 이벤트 발행을 Notify 와 Dispatch 로 가르고, 계정 삭제 정리를 거기에 건다

- 상태: 채택
- 날짜: 2026-10-15
- 관련: [ADR-0006](ADR-0006-account-recovery.md) 의 `Recovery`, [ADR-0007](ADR-0007-email-verification.md) 의 `Verification`

## 배경

`docs/tasks.md` 에 "만료된 `Recovery` 행 청소", "만료된 `Verification` 행 청소" 가 있었다.
둘 다 "탈퇴한 계정의 행은 남는다" 를 이유로 적었는데, 들여다보니 문제가 그보다 컸다.

`DeleteUser` 는 `m.db.Delete(User{Name: username})` 한 줄이었다. 그래서 계정을 지워도
`tokens`, `recoveries`, `verifications`, `assigns`, `members` 행이 모두 남았다.
**`tokens` 가 남는 것이 가장 나쁘다.** 같은 이름으로 다시 가입하면 죽은 계정의
비밀번호와 passkey 가 새 계정에 그대로 붙는다.

정리를 이벤트로 붙이려고 보니 `pubsub.Hub` 가 그에 맞지 않았다.
`Publish` 는 핸들러를 호출자의 goroutine 에서 동기로 돌리면서 에러는
`logrus.Debug` 로 버렸다. 호출자를 기다리게 하면서 결과는 주지 않는, 두 방식의
나쁜 쪽만 모은 모양이었다.

## 결정

### 1. 발행을 `Notify` 와 `Dispatch` 둘로 가른다

```go
Notify(ctx context.Context, detail any)          // async, 에러는 로그로만
Dispatch(ctx context.Context, detail any) error  // sync, 핸들러 에러를 모아 돌려줌
```

갈림길은 "부르는 쪽이 결과를 쓰는가" 하나다.

- `Notify` — 게시물 알림(`posts.EventPostCreated`)처럼 SSE 로 흘려보내면 그만인 것.
  응답을 기다릴 이유가 없고, 구독자가 실패해도 게시물 작성이 실패할 일은 아니다
- `Dispatch` — 계정 삭제 정리처럼 실패를 삼키면 넣은 의미가 없는 것

한 가지로 묶으면 둘 중 하나는 반드시 어긋난다. `Publish` 라는 이름은 지웠다.
남겨 두면 어느 쪽이었는지가 호출측에서 보이지 않는다.

`deliver` 는 핸들러 하나가 실패해도 나머지를 멈추지 않고 `errors.Join` 으로 모은다.
정리 핸들러가 여럿일 때 앞선 하나 때문에 뒤쪽이 통째로 빠지면 치우려던 것이 더 남는다.

### 2. `Notify` 는 큐를 거쳐 순서를 지킨다

이벤트마다 goroutine 을 띄우면 거의 동시에 올라온 게시물 둘의 순서가 SSE 에서
뒤바뀔 수 있다. 그래서 발행을 채널에 넣고 goroutine 하나가 차례로 꺼낸다.
`datastruct.DynamicChan` 이 받는 쪽에서 무한히 쌓아 주므로 넣는 쪽은 막히지 않는다.

큐는 닫지 않는다. 닫아 두면 종료 중에 들어온 `Notify` 가 닫힌 채널에 써서 죽는다.
대신 꺼내는 쪽이 `ctx.Done()` 에서 멈추고, `Notify` 도 같은 신호를 보고 넣기를 그만둔다.

`Notify` 가 쓰는 이벤트 context 는 호출자의 것이 아니라 `NewHub` 이 받은
프로세스 context 다. 요청 context 를 그대로 실으면 응답이 끝나는 순간 취소되어
뒤늦게 도는 핸들러가 죽은 context 를 받는다. `jobs.Manager.processContext` 와 같은 수다.

### 3. 계정 삭제 정리는 삭제 **뒤에** `Dispatch` 로 보낸다

```go
func (m *Manager) DeleteUser(username string) error {
	user, err := m.GetUser(username)
	...
	m.db.Delete(User{Name: username})
	return m.events.Dispatch(m.processContext, EventUserDeleted{User: *user})
}
```

**버린 선택지: 삭제 전에 보내기.** 핸들러가 모두 성공해야 `users` 행을 지우는
모양이다. 중간에 실패하면 계정이 남아 다시 시도할 수 있다는 장점이 있다.
고르지 않았다. 그 사이에 토큰만 지워지고 계정은 살아 있는 구간이 생긴다.
그 계정은 로그인할 수 없는데 존재하기는 해서, 사용자 눈에는 "지워지지도 않고
들어갈 수도 없는" 상태가 된다. 지워진 것이 확실한 쪽이 낫다.

**버린 선택지: 트랜잭션을 이벤트에 실어 보내기.** `Event.Context` 에 gorm 트랜잭션을
담으면 전부 되거나 전부 안 되게 만들 수 있다. 고르지 않았다. 모든 핸들러가 gorm 과
같은 DB 를 써야 하고, `pubsub` 이 영속계층을 알게 된다.

그래서 **`Dispatch` 의 에러는 "되돌려라" 가 아니라 "계정은 없어졌고 딸린 행이 남았다"
는 뜻이다.** 같은 이름으로 다시 `DeleteUser` 를 불러 마저 치울 수 있다.
이벤트가 트랜잭션 밖이라는 것은 에러를 받아도 달라지지 않는다.

### 4. auth 자신의 행도 핸들러로 치운다

`userCleanup` 이 `EventUserDeleted` 를 받아 `tokens`, `recoveries`, `verifications`,
`assigns`, `members` 를 한 트랜잭션에서 지운다. `auth.New` 가 자기를 등록한다.

자기 이벤트를 자기가 받는 모양이고, 지금 치울 것이 전부 `auth` 안에 있어서
`DeleteUser` 안에서 바로 지우는 쪽이 짧다. 그래도 핸들러로 둔다. 치우는 쪽을
`EventUserDeleted` 라는 한 자리에 모아 두면, 나중에 다른 백엔드가 사용자에 딸린
데이터를 갖게 됐을 때 `auth` 가 그 백엔드를 알지 않고도 붙는다.

이 때문에 `auth.New` 가 `ctx` 와 `*pubsub.Hub` 를 받게 됐다.
`posts.New`, `jobs.New` 와 같은 모양이 됐다.

### 5. 만료된 행은 조회할 때 지운다

`ValidateRecovery` 와 `ConfirmVerification` 이 만료된 행을 만나면 그 자리에서 지운다.
주기 작업을 돌리는 장치가 없고(`internal/workers` 는 fan-out 파이프라인이다),
만들 만큼 급하지도 않다. 사용자당 한 행이고 재발급하면 덮어쓴다.

아무도 링크를 누르지 않으면 행은 그대로 남는다. 그 계정이 지워질 때 4 가 치운다.

## 결과

- 계정을 지우면 딸린 행이 같이 사라진다. 같은 이름으로 다시 가입해도 옛 비밀번호로
  들어올 수 없다
- 호출측이 발행 방식을 고르게 됐다. `Notify` 는 막지 않고, `Dispatch` 는 알려준다
- `pubsub.Router` 는 아직 `Publish` 하나뿐이다. 쓰는 곳이 없어 건드리지 않았다
- 탈퇴 경로(`DeleteUser` 를 부르는 API)는 여전히 없다. `docs/tasks.md` 에 남겼다
