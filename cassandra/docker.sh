docker network create backend
docker run -d --name cassandra -p 7000:7000 -p 9042:9042 -v cassandra_data:/var/lib/cassandra -e CASSANDRA_START_RPC=true --network backend cassandra