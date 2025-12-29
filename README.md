# Yakusoku

Consumer-driven な契約テスト CLI ツール。Pact Specification v3/v4 互換。

## 特徴

- **Consumer SDK** - Consumer テストで契約の期待値を定義
- **Provider 検証** - 契約ファイルに対して Provider API を検証
- **Pact v3 互換** - Pact Specification v3 と完全互換
- **CLI ツール** - コマンドラインから契約を管理・検証

## インストール

### ソースからビルド

```bash
git clone https://github.com/jt-chihara/yakusoku.git
cd yakusoku
make build
```

バイナリは `bin/` ディレクトリに生成されます。

## クイックスタート

### 1. Consumer 契約を定義 (Go SDK)

```go
package main

import (
    "net/http"
    "testing"

    "github.com/jt-chihara/yakusoku/sdk/go/yakusoku"
)

func TestUserServiceClient(t *testing.T) {
    pact := yakusoku.NewPact(yakusoku.Config{
        Consumer: "OrderService",
        Provider: "UserService",
        PactDir:  "./pacts",
    })
    defer pact.Teardown()

    // インタラクションを定義
    pact.
        Given("user 1 exists").
        UponReceiving("a request for user 1").
        WithRequest(yakusoku.Request{
            Method: "GET",
            Path:   "/users/1",
        }).
        WillRespondWith(yakusoku.Response{
            Status: 200,
            Headers: map[string]string{
                "Content-Type": "application/json",
            },
            Body: map[string]interface{}{
                "id":   1,
                "name": "John Doe",
            },
        })

    // 実際のクライアントコードで検証
    err := pact.Verify(func() error {
        resp, err := http.Get(pact.ServerURL() + "/users/1")
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        // クライアントコードをここに記述
        return nil
    })

    if err != nil {
        t.Fatal(err)
    }
}
```

このテストを実行すると、`./pacts/orderservice-userservice.json` に契約ファイルが生成されます。

### 2. Provider を検証

```bash
yakusoku verify \
  --provider-base-url http://localhost:8080 \
  --pact-file ./pacts/orderservice-userservice.json
```

Provider States を使用する場合:

```bash
yakusoku verify \
  --provider-base-url http://localhost:8080 \
  --pact-file ./pacts/orderservice-userservice.json \
  --provider-states-setup-url http://localhost:8080/provider-states
```

## CLI コマンド

### verify

契約ファイルに対して Provider API を検証します。

```bash
yakusoku verify [flags]

フラグ:
  --provider-base-url string           Provider API のベース URL (必須)
  --pact-file string                   契約ファイルのパス (必須)
  --provider-states-setup-url string   Provider States セットアップ URL
  --verbose                            詳細出力を表示
```

### version

バージョン情報を表示します。

```bash
yakusoku version
```

## 契約ファイルフォーマット

Yakusoku は Pact Specification v3 フォーマットを使用します:

```json
{
  "consumer": { "name": "OrderService" },
  "provider": { "name": "UserService" },
  "interactions": [
    {
      "description": "a request for user 1",
      "providerState": "user 1 exists",
      "request": {
        "method": "GET",
        "path": "/users/1"
      },
      "response": {
        "status": 200,
        "headers": {
          "Content-Type": "application/json"
        },
        "body": {
          "id": 1,
          "name": "John Doe"
        }
      }
    }
  ],
  "metadata": {
    "pactSpecification": { "version": "3.0.0" }
  }
}
```

## Provider States

Provider States を使用すると、検証前にテストデータをセットアップできます。以下の形式の POST リクエストを受け付けるエンドポイントを実装してください:

```json
{
  "state": "user 1 exists",
  "params": { "userId": 1 }
}
```

実装例:

```go
http.HandleFunc("/provider-states", func(w http.ResponseWriter, r *http.Request) {
    var state struct {
        State  string                 `json:"state"`
        Params map[string]interface{} `json:"params"`
    }
    json.NewDecoder(r.Body).Decode(&state)

    switch state.State {
    case "user 1 exists":
        // テストデータベースに user 1 をセットアップ
    }

    w.WriteHeader(http.StatusOK)
})
```

## 開発

### 必要条件

- Go 1.24+

### ビルド

```bash
make build
```

### テスト

```bash
make test
```

### Lint

```bash
make lint
```

### 全チェック

```bash
make all  # lint, test, build
```

## プロジェクト構成

```
.
├── cmd/
│   └── yakusoku/          # CLI エントリーポイント
├── internal/
│   ├── cli/               # CLI コマンド
│   ├── contract/          # 契約の型、パーサー、バリデーター、ライター
│   ├── matcher/           # マッチングルール
│   ├── mock/              # モック HTTP サーバー
│   └── verifier/          # Provider 検証
├── sdk/
│   └── go/yakusoku/       # Go SDK
└── tests/
    ├── unit/              # ユニットテスト
    └── integration/       # 統合テスト
```

## ライセンス

MIT
