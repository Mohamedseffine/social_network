#!/bin/bash

API_URL="http://localhost:8080/api/posts/create_comment"

if [ ! -f session_token.txt ]; then
  echo "Error: session_token.txt not found"
  exit 1
fi

TOKEN=$(cat session_token.txt | tr -d '[:space:]')

POST_ID=19
COMMENT="create new comment use script!"

curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -b "session_token=$TOKEN" \
  -d '{
    "post_id": '"$POST_ID"',
    "content": "'"$COMMENT"'"
  }' | jq
