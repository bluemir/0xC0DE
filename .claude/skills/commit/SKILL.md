---
name: commit
description: Stage changed files and create a git commit
disable-model-invocation: true
allowed-tools:
  - Bash(git status)
  - Bash(git diff:*)
  - Bash(git log:*)
  - Bash(git add:*)
  - Bash(git commit:*)
  - Read
  - Glob
  - Grep
---

# Commit Skill

변경사항을 스테이징하고 커밋을 생성합니다. 인자가 주어지면 커밋 메시지 작성 시 참고합니다.

## 절차

1. **현재 상태 파악**: 아래 명령을 **병렬**로 실행합니다.
   - `git status` (untracked 파일 포함, `-uall` 플래그는 사용 금지)
   - `git diff` (staged + unstaged 변경사항)
   - `git log --oneline -5` (최근 커밋 메시지 스타일 확인)

2. **변경사항 분석**: diff 내용을 읽고, 변경의 성격을 파악합니다.

3. **커밋 메시지 작성**: 이 프로젝트의 Conventional Commit 스타일을 따릅니다.
   - 형식: `type: 간결한 설명`
   - type 종류: `feature`, `fix`, `refactor`, `docs`, `test`, `chore`, `style`, `perf`, `ci`, `build`
   - 설명은 변경의 "why"에 초점을 맞춥니다
   - 본문이 필요하면 빈 줄 후 상세 내용을 추가합니다
   - 인자(`$ARGUMENTS`)가 주어졌으면 커밋 메시지 작성 시 참고합니다

4. **스테이징 및 커밋**:
   - 민감 파일(.env, credentials, 토큰 등)은 절대 스테이징하지 않습니다
   - `git add`는 특정 파일을 지정하여 실행합니다 (`git add -A` 금지)
   - 커밋 메시지는 HEREDOC 형식으로 전달합니다:
     ```
     git commit -m "$(cat <<'EOF'
     type: 커밋 메시지

     EOF
     )"
     ```

5. **결과 확인**: `git status`로 커밋 성공을 확인합니다.

## 주의사항

- pre-commit hook 실패 시 `--amend`가 아닌 **새 커밋**을 생성합니다
- `--no-verify` 등 hook 우회 옵션을 사용하지 않습니다
- push는 하지 않습니다 (사용자가 명시적으로 요청한 경우에만)
- 변경사항이 없으면 빈 커밋을 만들지 않습니다
