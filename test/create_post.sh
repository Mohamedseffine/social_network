# I have pushed the changes to GitHub. Here are the instructions to test the new create post endpoint:

# Start the server from the backend directory. This will run the new migrations for the posts and comments tables.

# Log in as any user and save their cookie:

curl -c cookie.txt -X POST -H "Content-Type: application/json" -d '{
    "email": "usera@test.com", "password": "password"
}' http://localhost:8081/api/login
# Create a new post. You should get a 201 Created response containing the new post data.

curl -i -b cookie.txt -X POST -H "Content-Type: application/json" -d '{
    "content": "This is my first post!",
    "privacy": "public"
}' http://localhost:8081/api/posts
# Please let me know the result.