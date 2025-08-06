# Start the server from the backend directory.

# Get the list of users UserA is following. The response should be a JSON array containing UserB's public information.

curl http://localhost:8081/api/users/2/following
# Get the list of UserB's followers. The response should be a JSON array containing UserA's public information.

curl http://localhost:8081/api/users/2/followers