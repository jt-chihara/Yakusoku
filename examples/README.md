# Yakusoku サンプルアプリケーション

契約テストの動作を確認するためのサンプルアプリケーションです。

## 構成

```
examples/
├── userservice/     # Provider (UserService)
│   └── main.go      # ユーザー情報を提供する API
├── orderservice/    # Consumer (OrderService)
│   ├── main.go      # 注文サービス (UserService を利用)
│   └── consumer_test.go  # 契約テスト
└── README.md
```

## 使い方

### 1. Consumer テストを実行 (契約ファイル生成)

```bash
cd examples/orderservice
go test -v ./...
```

これにより `pacts/orderservice-userservice.json` が生成されます。

### 2. 生成された契約を確認

```bash
yakusoku show --pact-file ../../pacts/orderservice-userservice.json
```

### 3. Provider を起動

```bash
go run ./examples/userservice
```

### 4. Provider を契約に対して検証

別ターミナルで:

```bash
yakusoku verify \
  --provider-base-url http://localhost:8080 \
  --pact-file pacts/orderservice-userservice.json \
  --provider-states-setup-url http://localhost:8080/provider-states
```

### 5. (オプション) 両方のサービスを起動して統合テスト

ターミナル 1:
```bash
go run ./examples/userservice
```

ターミナル 2:
```bash
go run ./examples/orderservice
```

ターミナル 3:
```bash
curl http://localhost:8081/orders/1
```

## 契約テストのフロー

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  OrderService   │     │    Contract     │     │  UserService    │
│   (Consumer)    │     │     File        │     │   (Provider)    │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         │  1. Consumer Test     │                       │
         │─────────────────────▶│                       │
         │   (Mock Server)       │                       │
         │                       │                       │
         │  2. Generate Contract │                       │
         │─────────────────────▶│                       │
         │                       │                       │
         │                       │  3. Verify Provider   │
         │                       │──────────────────────▶│
         │                       │                       │
         │                       │  4. Verification OK   │
         │                       │◀──────────────────────│
         │                       │                       │
```

## API

### UserService (Provider)

| Method | Path | Description |
|--------|------|-------------|
| GET | /users | ユーザー一覧取得 |
| GET | /users/{id} | ユーザー取得 |
| POST | /provider-states | Provider State セットアップ |

### OrderService (Consumer)

| Method | Path | Description |
|--------|------|-------------|
| GET | /orders/{id} | 注文取得 (UserService からユーザー情報を取得) |
