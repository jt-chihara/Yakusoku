# frozen_string_literal: true

# Consumer contract test for UserService
#
# Run with: bundle exec rspec spec/contracts/user_service_spec.rb
#
# This test generates a contract file at ./pacts/orderservice-userservice.json

require "spec_helper"

RSpec.describe "UserService Contract" do
  let(:pact) do
    Yakusoku::Pact.new(
      consumer: "OrderService",
      provider: "UserService",
      pact_dir: "./pacts"
    )
  end

  after { pact.teardown }

  it "defines all interactions with UserService" do
    # Define all expected interactions

    # 1. GET /users/:id - success case
    pact
      .given("user 1 exists")
      .upon_receiving("a request to get user 1")
      .with_request(method: "GET", path: "/users/1")
      .will_respond_with(
        status: 200,
        headers: { "Content-Type" => "application/json" },
        body: { id: 1, name: "John Doe", email: "john@example.com" }
      )

    # 2. GET /users/:id - not found case
    pact
      .given("user 999 does not exist")
      .upon_receiving("a request to get non-existent user")
      .with_request(method: "GET", path: "/users/999")
      .will_respond_with(
        status: 404,
        headers: { "Content-Type" => "application/json" },
        body: { error: "User not found" }
      )

    # 3. POST /users - create user
    pact
      .given("no users exist")
      .upon_receiving("a request to create a user")
      .with_request(
        method: "POST",
        path: "/users",
        headers: { "Content-Type" => "application/json" },
        body: { name: "Jane Doe", email: "jane@example.com" }
      )
      .will_respond_with(
        status: 201,
        headers: { "Content-Type" => "application/json" },
        body: { id: 2, name: "Jane Doe", email: "jane@example.com" }
      )

    # 4. GET /users/:id/orders
    pact
      .given("user 1 has orders")
      .upon_receiving("a request to get user 1's orders")
      .with_request(method: "GET", path: "/users/1/orders")
      .will_respond_with(
        status: 200,
        headers: { "Content-Type" => "application/json" },
        body: [
          { id: 101, total: 99.99 },
          { id: 102, total: 149.99 }
        ]
      )

    # Verify all interactions with actual client code
    pact.verify do |mock_server_url|
      client = UserServiceClient.new(base_url: mock_server_url)

      # Test 1: Get existing user
      user = client.get_user(1)
      expect(user[:id]).to eq(1)
      expect(user[:name]).to eq("John Doe")
      expect(user[:email]).to eq("john@example.com")

      # Test 2: Get non-existent user
      expect { client.get_user(999) }.to raise_error(UserServiceClient::NotFoundError)

      # Test 3: Create user
      new_user = client.create_user(name: "Jane Doe", email: "jane@example.com")
      expect(new_user[:id]).to eq(2)
      expect(new_user[:name]).to eq("Jane Doe")

      # Test 4: Get user orders
      orders = client.get_user_orders(1)
      expect(orders.length).to eq(2)
      expect(orders[0][:id]).to eq(101)
      expect(orders[1][:total]).to eq(149.99)
    end
  end
end
