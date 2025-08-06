# Register User A
curl -X POST -H "Content-Type: application/json" -d '{
    "email": "usera@test.com", "password": "password", "first_name": "User", "last_name": "A", "date_of_birth": "2000-01-01"
}' http://localhost:8081/api/register

# Register User B
curl -X POST -H "Content-Type: application/json" -d '{
    "email": "userb@test.com", "password": "password", "first_name": "User", "last_name": "B", "date_of_birth": "2000-01-01"
}' http://localhost:8081/api/register