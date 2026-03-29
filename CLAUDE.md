## 행동 규칙

- 애매하거나 결정이 필요한 사항은 질문을 반드시 할것
- 의도나 용어는 추측하지 말고 물어볼 것
- 명확하지기 전에는 계속 해서 재질문 할것

## Code Style

### Frontend

- CSS 와 HTML element 를 최소화 한다.
- SPA 로 구현하지 말것
- alert(), confirm() 은 사용하지 말것
- 최신 ECMAScript 및 웹 표준을 사용할것

### Backend

- 과도한 추상화를 하지 않는다.
	- interface 는 반드시 필요하기 전에는 도입하지 않는다.
	- Depandancy Injection 은 최소화 한다.



## 문서

### Roadmap

- 로드맵은 vivid 가 제공하는 기능과 앞으로 구현될 내용을 모두 담고 있어야 한다.
- 로드맵은 project 가 진행 되면서 조금씩 변경 될수 있다.
	- 구현 도중 로드맵의 변경이 필요한 사항은 한번더 확인하고 진행 한다.

### ADR (Architecture Decision Records)

- 아키텍처/기술 선택에 trade-off가 있는 결정을 할 때 `docs/adr/`에 ADR을 작성한다
- 파일명: `ADR-NNNN-제목.md` (예: `ADR-0001-use-redis-for-caching.md`)
- 기존 ADR이 있으면 번호를 이어서 채번한다
- 버그 수정, 단순 리팩토링, 선택지가 하나뿐인 경우는 작성하지 않는다

