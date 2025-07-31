#!/bin/bash

API_URL="http://localhost:8080/api/users"
TOKEN=$(cat session_token.txt)

curl -s -X GET "$API_URL" \
  -H "Content-Type: application/json" \
  -b "session_token=$TOKEN" | jq
