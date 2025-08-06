# You are right, my apologies. I will provide the test instructions now. I have pushed the changes to GitHub.

# Here is how you can test the new update profile endpoint:

# Start the server from the backend directory.

# Log in as an existing user (e.g., usera@test.com) and save their cookie:

curl -c cookieA.txt -X POST -H "Content-Type: application/json" -d '{
    "email": "usera@test.com", "password": "password"
}' http://localhost:8081/api/login
# Update the user's profile. This example changes the user's first name and sets their profile to public. You should get a 200 OK response.

curl -i -b cookieA.txt -X PUT -H "Content-Type: application/json" -d '{
    "first_name": "User",
    "profile_type": "private"
}' http://localhost:8081/api/profile
# Verify the change by fetching the user's profile again (UserA has ID 1). The response should show the new first name and profile type.

curl -b cookieA.txt http://localhost:8081/api/users/1/profile
# Please let me know the results.