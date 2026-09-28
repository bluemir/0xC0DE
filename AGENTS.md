## 행동 규칙

- 애매하거나 결정이 필요한 사항은 질문을 반드시 할것
- 의도나 용어는 추측하지 말고 물어볼 것
- 명확하지기 전에는 계속 해서 재질문 할것

## Code Style

### Frontend

- CSS 와 HTML element 를 최소화 한다.
- SPA 로 구현하지 말것
- `alert()`, `confirm()` 은 사용하지 말것
	- 대신 `<dialog>` 과 `popover` attribute 를 사용할것
- 최신 ECMAScript 및 웹 표준을 사용할것
- 필요하면 https://github.com/VoltAgent/awesome-design-md 를 참조 할수 있다.

### Backend

- 과도한 추상화를 하지 않는다.
	- interface 는 반드시 필요하기 전에는 도입하지 않는다.
		- 예외: 대형 구조체(Manager 등)의 전체 공개 API 색인/가이드 제공 및 컴파일 타임 일치 검증(`var _ I... = (*...)(nil)`)을 위한 인터페이스는 허용한다.
	- Depandancy Injection 은 최소화 한다.
	- 단순 중복이 있다고 helper 를 만들지 않는다.
- 코드는 의도에 따라 작성한다.
	- 변수명은 과도하게 축약하지 않는다.(eg 한글자 변수명)
		- 예외 loop 순환자는 한글자 변수명을 쓸수 있다.
	- 함수의 parameter나 return 값에 flag 를 무분별 하게 넣지 않는다.

## 문서

### Roadmap

- 로드맵은 제공하는 기능과 앞으로 구현될 내용을 모두 담고 있어야 한다.
	- `docs/roadmap.md` 에 기록 한다.
- 로드맵은 project 가 진행 되면서 조금씩 변경 될수 있다.
	- 구현 도중 로드맵의 변경이 필요한 사항은 한번더 확인하고 진행 한다.

### TODO

- 실행 계획이나 추후 개선점 등 해야 할일은 `docs/tasks.md` 에 기록한다.
- `- [ ] 내용` 처럼 checklist 형태로 기록 한다.
- 하나의 list가 되도록 기록한다.
	- 세부 사항은 하위 항목으로 기록 한다.
- 작업이 완료되면 `[x]` 로 체크한다.
	- 작업의 결과를 기록하지 않는다. 결과는 ADR 이나 다른 문서에 기록 한다.
	- 후속으로 할일만 새 항목으로 작성 한다.
- 끝낸 항목중 오래되거나 장기 추적이 필요 없는 건은 주기적으로 `docs/done.md` 로 옮겨진다.
	- 옯기는 작업은 별도로 명시적으로 진행 하며, 임의로 진행 하지 않는다.

예시

```
- [ ] example 1
- [ ] example 2
	- descrition 1
	- descrition 2
	- descrition 3
- epic example
	- [ ] job1
	- [ ] job2
- [x] done
```


### Issues

- 남의 코드(의존 라이브러리·터미널) 에서 난 문제는 `docs/issues/` 에 적는다
- 파일명: `NNNN-짧은-제목.md`
- 재현 절차와 "우리 코드가 아니라는 근거"를 반드시 남긴다
- 위에 보고하지 않기로 한 것도 적는다
	- 다음에 같은 증상을 처음부터 다시 파는 것이 가장 비싸다

### Reference

- 참고할만만 Article, 영상 등의 요약, 정리는 `docs/references/` 아래에 작성한다.
- 파일명: `짧은-제목.md`

### ADR (Architecture Decision Records)

- 아키텍처/기술 선택에 trade-off가 있는 결정을 할 때 `docs/adr/`에 ADR을 작성한다
- 파일명: `ADR-NNNN-제목.md` (예: `ADR-0001-use-redis-for-caching.md`)
- 기존 ADR이 있으면 번호를 이어서 채번한다
- 버그 수정, 단순 리팩토링, 선택지가 하나뿐인 경우는 작성하지 않는다

