curl -i -b cookieB.txt -X POST -H "Content-Type: application/json" -d '{
    "action": "accept"
}' http://localhost:8081/api/follow-requests/1