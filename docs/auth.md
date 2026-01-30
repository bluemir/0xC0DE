# 인증(Authentication) 및 인가(Authorization)

이 문서는 시스템의 인증 및 인가 처리 방식에 대해 설명합니다.

## 개요

시스템은 역할 기반 접근 제어(RBAC, Role-Based Access Control)를 사용하여 리소스에 대한 접근 권한을 관리합니다.

## 주요 개념

### Role (역할)

Role은 사용자가 수행할 수 있는 작업의 집합을 정의합니다. `internal/server/backend/auth/role.go`에 정의되어 있습니다.

```go
type Role struct {
    Name  string
    Rules []Rule
}
```

### Rule (규칙)

Rule은 특정 리소스에 대해 허용된 작업(Verb)을 정의합니다.

```go
type Rule struct {
    Verbs      []Verb      // 허용된 동작 목록 (예: "create", "delete")
    Selector   KeyValues   // 대상 리소스 선택자
    Conditions []Condition // 추가 조건
}
```

### Verb (동작)

Verb는 리소스에 대해 수행할 수 있는 동작을 나타냅니다. (예: `get`, `list`, `create`, `update`, `delete` 등)

> [!IMPORTANT]
> **와일드카드 처리 변경 사항**
>
> 기존에는 `"*"` 문자열을 Verb로 설정하면 모든 동작을 허용하는 와일드카드로 작동했으나, 이 기능은 제거되었습니다.
> 현재 모든 동작을 허용하려면 **빈 리스트(`[]Verb{}`)**를 사용해야 합니다.

### Selector (선택자)

Selector는 Rule이 적용될 리소스를 선택하는 데 사용됩니다. Key-Value 쌍으로 정의됩니다.

### Condition (조건)

Condition은 Rule이 적용되기 위해 만족해야 하는 추가적인 제약 조건입니다.

## 사용 예시

### 특정 리소스에 대한 읽기 전용 권한

```go
Rule{
    Verbs: []Verb{"get", "list"},
    Selector: KeyValues{"kind": "pod"},
}
```

### 모든 동작 허용 (슈퍼 유저)

기존의 `"*"` 대신 빈 리스트를 사용합니다.

```go
Rule{
    Verbs: []Verb{}, // 비워두면 모든 동작 허용
    Selector: KeyValues{}, // 비워두면 모든 리소스 선택
}
```
