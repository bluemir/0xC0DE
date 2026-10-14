# Go 로 SVG 를 아이콘 폰트로 묶기

자체 SVG 아이콘(수십 개 규모)을 웹폰트로 굽는 툴을 Go 로 만들 수 있는지 조사한 결과.
2026-10-14 기준. 착수하지 않기로 했고, 할 때 처음부터 다시 찾지 않으려고 남긴다.

## 결론

Go 로 SVG → 아이콘 폰트를 통째로 해주는 라이브러리는 **없다**.
이 일을 하는 도구는 전부 Node(`svgtofont`, `fantasticon`, `svgicons2svgfont`)이거나 Dart(`icon_font_generator`)다.

다만 제일 부담스러워 보였던 **sfnt 테이블 직접 작성은 라이브러리가 해결해 준다**.
직접 할 일은 기하 변환과 조립으로 좁혀진다.

## 왜 폰트인가 (SVG sprite 가 아니라)

아이콘은 `assets/src/js/components/icon/icon.js` 의 `c-icon` 이 **shadow DOM** 안에서 그린다.

- `@font-face` 는 document 에만 등록되고 shadow root 안에서는 무시된다.
  지금은 `assets/html-templates/snippet/header.html` 이 document 레벨에 `fonts.css` 를 link 해서 동작한다.
  즉 폰트 방식은 현재 구조에 그대로 맞는다.
- SVG sprite 는 Chromium 이 `<use href="외부파일#id">` 를 지원하지 않아
  `c-icon` 이 직접 fetch 해서 인라인으로 박아야 한다. 구조 변경이 따라온다.

## 후보 라이브러리

### 폰트 쓰기

| 라이브러리 | 판정 |
|---|---|
| [`seehuhn.de/go/sfnt`](https://github.com/seehuhn/go-sfnt) | **채택 후보.** 읽기/쓰기 모두 한다 |
| [`golang.org/x/image/font/sfnt`](https://pkg.go.dev/golang.org/x/image/font/sfnt) | 읽기 전용. `WriteSourceTo` 는 원본 바이트 복사라 새 glyf 를 못 만든다 |
| [`ConradIrwin/font/sfnt`](https://pkg.go.dev/github.com/ConradIrwin/font/sfnt) | `WriteOTF` 로 직렬화는 되지만 **이미 글리프를 가진 Font 객체**가 전제다. 조립 API 가 아니다 |
| [`tdewolff/font`](https://github.com/tdewolff/font) | WOFF/WOFF2/EOT → SFNT 변환기. 읽기 방향만 |

`seehuhn.de/go/sfnt` 는 `glyf`/`loca`/`cmap`/`hmtx`/`head`/`maxp`/`name`/`OS/2`/`post` 가
각각 서브패키지로 있고 전부 write 를 한다. 글리프는 바이너리가 아니라 좌표로 조립한다.

```go
su := &glyf.SimpleUnpacked{
	Contours: []glyf.Contour{{
		{X: 0, Y: 0, OnCurve: true},
		{X: 500, Y: 0, OnCurve: true},
		{X: 250, Y: 700, OnCurve: false},
	}},
}
glyph := su.AsGlyph()
```

`Font.Write(w)` 로 뱉는다. `AsGlyph()` 가 bounding box 를 계산해 주는지는 문서에 없으니 확인이 필요하다.

### WOFF2 쓰기

순수 Go 네이티브 인코더는 없다. 두 갈래가 있다.

- [`pgaskin/go-woff2`](https://github.com/pgaskin/go-woff2) — `Encode(ttf []byte, params)`.
  구글 woff2 를 wasm 으로 빌드해 `wasm2go` 로 옮긴 것. **cgo 는 없지만** 손으로 쓴 Go 는 아니다.
- 직접 쓰기 — WOFF2 는 glyf/loca 에 **null transform(version 3)** 을 허용한다.
  변환 없이 `header + table directory(base-128 varint) + brotli 압축`만 해도 유효하다.
  brotli 는 `andybalholm/brotli` 가 순수 Go 다.
  full transform 대비 파일이 조금 커지는 대신 분량이 훨씬 적다.

## 정한 방향: PUA codepoint

아이콘을 지정하는 방식은 두 가지인데 **PUA codepoint + CSS 클래스**(FontAwesome 방식)로 간다.

- PUA — `U+E000~` 에 할당하고 `.icon-home::before { content: "\e001" }`.
  `cmap` 만 쓰면 돼서 구현이 단순하다.
- ligature — Material Symbols 방식. `c-icon` 의 Material 분기와 모양이 같아 전환은 쉽지만,
  `seehuhn.de/go/sfnt` 의 `gtab/builder` 가 문서화돼 있지 않아 소스를 뚫어야 한다.

난이도 차이가 실제로 크다는 점이 판단 근거다.

## 입력 제약: fill path 만 받는다

stroke → outline 변환(선 굵기를 윤곽선으로 펴고 join/cap/miter 처리)은
폰트 테이블 작성보다 어렵고, 어떤 라이브러리도 대신 해주지 않는다.

아이콘 SVG 는 Figma/Illustrator 에서 "Outline Stroke" 한 번이면 fill 로 떨어진다.
입력 규칙으로 두고 stroke 가 있으면 에러를 내는 쪽이 낫다.

## 남는 작업

| 단계 | 누가 |
|---|---|
| SVG `d` 속성 파싱 | 직접. 아이콘은 보통 `<path>` 하나라 작다 |
| cubic → quadratic 변환 | 직접. glyf 는 2차 베지어만 받는다. midpoint 분할 |
| viewBox → em square 정규화, y축 반전 | 직접 |
| glyf/cmap/hmtx/name/OS2 작성 | `seehuhn.de/go/sfnt` |
| TTF → WOFF2 | `pgaskin/go-woff2`, 또는 null transform 직접 |
| `@font-face` 와 클래스 CSS 생성 | `text/template` |

## 둘 곳

`scripts/tools/svg-font/` + **자체 `go.mod`**.
`scripts/tools/make-select/` 가 이미 같은 형태다 — repo 안에 있으면서 모듈 경로는 독립이라
루트 모듈과 의존성이 섞이지 않고, 자라면 `git subtree split` 으로 뗄 수 있다.

처음부터 별도 repo 를 파면 태깅·릴리스·루트 `go.mod` 의 버전 고정이 따라붙는데,
아직 아이콘이 한 장도 없는 단계에서는 이르다.

## 안 만들기로 한 선택지

- **npm 체인에 얹기** — `svgicons2svgfont` → `svg2ttf` → `ttf2woff2`, 또는 `fantasticon`.
  `package.json` 에 넣고 `scripts/makefile.d/web.mk` 에 규칙 한 줄이면 끝이라 가장 싸다.
  이미 `node_modules` + esbuild 파이프라인이 있어 결도 맞는다. Go 로 직접 하기로 해서 택하지 않았다.
- **폰트 없이 인라인 SVG** — `c-icon` 이 SVG 를 fetch 해서 shadow DOM 에 인라인.
  빌드 단계가 없고 `currentColor`·다색이 되지만, 아이콘이 수십 개 규모면 폰트가 낫다.
