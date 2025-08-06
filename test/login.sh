curl -c cookieA.txt -X POST -H "Content-Type: application/json" -d '{
    "email": "usera@test.com", "password": "password"
}' http://localhost:8081/api/login

curl -c cookieB.txt -X POST -H "Content-Type: application/json" -d '{
    "email": "userb@test.com", "password": "password"
}' http://localhost:8081/api/login