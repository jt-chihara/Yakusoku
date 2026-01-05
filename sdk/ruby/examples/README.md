# Yakusoku Ruby SDK サンプル

契約テストの動作を確認するためのサンプルアプリケーションです。

## 構成

```
examples/
└── orderservice/           # Consumer (OrderService)
    ├── Gemfile
    ├── lib/
    │   └── user_service_client.rb  # UserService API クライアント
    ├── spec/
    │   ├── spec_helper.rb
    │   └── contracts/
    │       └── user_service_spec.rb  # 契約テスト
    └── pacts/              # 生成された契約ファイル
```

## 使い方

### 1. 依存関係をインストール

```bash
cd sdk/ruby/examples/orderservice
bundle install
```

### 2. Consumer テストを実行 (契約ファイル生成)

```bash
bundle exec rspec spec/contracts/user_service_spec.rb -f d
```

これにより `pacts/orderservice-userservice.json` が生成されます。

### 3. 生成された契約を確認

```bash
cat pacts/orderservice-userservice.json | jq .
```

または CLI で:

```bash
yakusoku show --pact-file pacts/orderservice-userservice.json
```

### 4. Provider を契約に対して検証

Provider (UserService) を起動した状態で:

```bash
yakusoku verify \
  --provider-base-url http://localhost:8080 \
  --pact-file pacts/orderservice-userservice.json \
  --provider-states-setup-url http://localhost:8080/provider-states
```

## テストケース

| テスト | 説明 |
|--------|------|
| GET /users/:id (存在) | ユーザー取得成功 |
| GET /users/:id (不存在) | 404 エラー |
| POST /users | ユーザー作成 |
| GET /users/:id/orders | ユーザーの注文一覧取得 |

## 契約テストのフロー

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  OrderService   │     │    Contract     │     │  UserService    │
│   (Consumer)    │     │     File        │     │   (Provider)    │
│     Ruby        │     │     JSON        │     │   Go/Rails/etc  │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         │  1. RSpec Contract    │                       │
         │     Test 実行         │                       │
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

## Rails での使い方

Rails アプリケーションで使う場合:

```ruby
# Gemfile
gem 'yakusoku', path: 'path/to/sdk/ruby/yakusoku'

# spec/contracts/user_service_spec.rb
require 'rails_helper'
require 'yakusoku/rspec'

RSpec.describe 'UserService Contract', type: :contract do
  let(:pact) do
    Yakusoku::Pact.new(
      consumer: Rails.application.class.module_parent_name,
      provider: 'UserService',
      pact_dir: Rails.root.join('pacts')
    )
  end

  after { pact.teardown }

  # テストを記述...
end
```

CI での実行:

```yaml
# .github/workflows/contract.yml
jobs:
  contract-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: ruby/setup-ruby@v1
        with:
          ruby-version: '3.2'
          bundler-cache: true
      - name: Run contract tests
        run: bundle exec rspec spec/contracts/ -f d
      - name: Upload pact files
        uses: actions/upload-artifact@v4
        with:
          name: pacts
          path: pacts/
```
