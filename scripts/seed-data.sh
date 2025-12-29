#!/bin/sh

BROKER_URL="${BROKER_URL:-http://broker:8080}"

echo "Waiting for broker to be ready..."
until curl -sf "${BROKER_URL}/pacts" > /dev/null 2>&1; do
  sleep 1
done
echo "Broker is ready!"

echo "Seeding sample contract..."
curl -sf -X POST "${BROKER_URL}/pacts/provider/UserService/consumer/OrderService/version/1.0.0" \
  -H "Content-Type: application/json" \
  -d @/data/sample-contract.json

echo ""
echo "Sample contract seeded successfully!"
echo "  Consumer: OrderService"
echo "  Provider: UserService"
echo "  Version: 1.0.0"
