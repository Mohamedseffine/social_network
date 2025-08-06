# I have pushed the changes to GitHub. Here are the instructions to test the new get posts feed endpoint. This test is a bit more involved as it requires setting up a specific scenario.

# Scenario: UserA makes two posts (one public, one "almost private"). UserB will view the feed, then follow UserA, then view the feed again to see the change.

# Start the server with a clean database.

# Register UserA and UserB.

# Log in as UserA and save the cookie:

curl -c cookieA.txt -X POST -H "Content-Type: application/json" -d '{"email": "usera@test.com", "password": "password"}' http://localhost:8081/api/login
# As UserA, create a public post:

curl -b cookieA.txt -X POST -H "Content-Type: application/json" -d '{"content": "Public post by UserA", "privacy": "public"}' http://localhost:8081/api/posts
# As UserA, create an "almost private" post:

curl -b cookieA.txt -X POST -H "Content-Type: application/json" -d '{"content": "Almost private post by UserA", "privacy": "almost_private"}' http://localhost:8081/api/posts
# Log in as UserB and save the cookie (UserB is not following UserA yet):

curl -c cookieB.txt -X POST -H "Content-Type: application/json" -d '{"email": "userb@test.com", "password": "password"}' http://localhost:8081/api/login
# As UserB, get the posts feed. The response should be a JSON array containing only the public post from UserA.

curl -b cookieB.txt http://localhost:8081/api/posts
# Now, have UserB follow UserA (and have UserA's profile be public so it's auto-accepted).

# As UserA, make profile public
curl -b cookieA.txt -X PUT -H "Content-Type: application/json" -d '{"profile_type": "public"}' http://localhost:8081/api/profile

# As UserB, follow UserA
curl -b cookieB.txt -X POST http://localhost:8081/api/users/1/follow
# As UserB, get the posts feed again. The response should now be a JSON array containing both posts from UserA.

curl -b cookieB.txt http://localhost:8081/api/posts
# Please let me know the results of steps 7 and 9.

