# Yakusoku Ruby SDK

Consumer-driven contract testing for Ruby/Rails. Pact Specification v3 互換。

## インストール

Gemfile に追加:

```ruby
gem 'yakusoku', path: 'path/to/sdk/ruby/yakusoku'
```

または RubyGems から（公開後）:

```ruby
gem 'yakusoku'
```

## 使い方

### 基本的な使い方

```ruby
require 'yakusoku'

pact = Yakusoku::Pact.new(
  consumer: 'OrderService',
  provider: 'UserService',
  pact_dir: './pacts'
)

# インタラクションを定義
pact
  .given('user 1 exists')
  .upon_receiving('a request for user 1')
  .with_request(method: 'GET', path: '/users/1')
  .will_respond_with(
    status: 200,
    headers: { 'Content-Type' => 'application/json' },
    body: { id: 1, name: 'Alice', email: 'alice@example.com' }
  )

# 検証
pact.verify do |mock_server_url|
  # クライアントコードでモックサーバーを呼ぶ
  response = Net::HTTP.get(URI("#{mock_server_url}/users/1"))
  user = JSON.parse(response)
  # アサーション
end
```

### RSpec との統合

```ruby
# spec/spec_helper.rb
require 'yakusoku/rspec'

# spec/contracts/user_service_spec.rb
RSpec.describe 'UserService Contract' do
  let(:pact) do
    Yakusoku::Pact.new(
      consumer: 'OrderService',
      provider: 'UserService',
      pact_dir: './pacts'
    )
  end

  after { pact.teardown }

  it 'returns user details' do
    pact
      .given('user 1 exists')
      .upon_receiving('a request for user 1')
      .with_request(method: 'GET', path: '/users/1')
      .will_respond_with(
        status: 200,
        headers: { 'Content-Type' => 'application/json' },
        body: { id: 1, name: 'Alice' }
      )

    pact.verify do |mock_server_url|
      client = UserServiceClient.new(base_url: mock_server_url)
      user = client.get_user(1)

      expect(user.id).to eq(1)
      expect(user.name).to eq('Alice')
    end
  end
end
```

### Rails での使い方

```ruby
# spec/contracts/payment_service_spec.rb
require 'rails_helper'
require 'yakusoku/rspec'

RSpec.describe 'PaymentService Contract', type: :contract do
  let(:pact) do
    Yakusoku::Pact.new(
      consumer: Rails.application.class.module_parent_name,
      provider: 'PaymentService',
      pact_dir: Rails.root.join('pacts')
    )
  end

  after { pact.teardown }

  describe 'POST /payments' do
    it 'creates a payment' do
      pact
        .given('valid payment credentials')
        .upon_receiving('a request to create a payment')
        .with_request(
          method: 'POST',
          path: '/payments',
          headers: { 'Content-Type' => 'application/json' },
          body: { amount: 1000, currency: 'JPY' }
        )
        .will_respond_with(
          status: 201,
          headers: { 'Content-Type' => 'application/json' },
          body: { id: 'pay_123', status: 'succeeded' }
        )

      pact.verify do |mock_server_url|
        client = PaymentClient.new(base_url: mock_server_url)
        result = client.create_payment(amount: 1000, currency: 'JPY')

        expect(result.status).to eq('succeeded')
      end
    end
  end
end
```

## API

### Pact

| メソッド | 説明 |
|---------|------|
| `given(state)` | Provider の状態を設定 |
| `upon_receiving(description)` | インタラクションの説明を設定 |
| `with_request(options)` | 期待するリクエストを設定 |
| `will_respond_with(options)` | 期待するレスポンスを設定 |
| `verify { \|url\| ... }` | モックサーバーを起動して検証 |
| `teardown` | リソースをクリーンアップ |
| `server_url` | モックサーバーの URL を取得 |

### Request Options

| オプション | 型 | 説明 |
|-----------|---|------|
| `method` | String | HTTP メソッド (GET, POST, etc.) |
| `path` | String | リクエストパス |
| `query` | Hash | クエリパラメータ |
| `headers` | Hash | リクエストヘッダー |
| `body` | Hash/String | リクエストボディ |

### Response Options

| オプション | 型 | 説明 |
|-----------|---|------|
| `status` | Integer | HTTP ステータスコード |
| `headers` | Hash | レスポンスヘッダー |
| `body` | Hash/String | レスポンスボディ |

## 契約ファイルの公開

生成された契約ファイルを Yakusoku Broker に公開:

```bash
curl -X POST http://broker:8080/pacts/provider/UserService/consumer/OrderService/version/1.0.0 \
  -H "Content-Type: application/json" \
  -d @./pacts/orderservice-userservice.json
```

## ライセンス

MIT
