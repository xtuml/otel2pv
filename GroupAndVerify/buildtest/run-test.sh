#!/bin/bash
set -e

docker compose -f docker-compose-buildtest.yml up -d

# Ample time for GroupAndVerify to start
sleep 5

# check the logs to see if the GroupAndVerify is running
docker compose -f docker-compose-buildtest.yml logs group-and-verify

# Publish message to the ConsumeTest queue using management API
echo "Publishing message to ConsumeTest queue..."
curl -X POST -H "content-type:application/json" http://localhost:15672/api/exchanges/%2f/amq.default/publish -d '{"vhost":"/","name":"amq.default","properties":{"delivery_mode":1,"headers":{"Content-Type":"application/json"}},"routing_key":"ConsumeTest","delivery_mode":"1","payload":"{\"treeId\":\"tree\",\"nodeId\":\"node1\",\"parentId\":\"\",\"childIds\":[\"node2\"],\"nodeType\":\"\",\"timestamp\":2,\"appJSON\":{}}","payload_encoding":"string","headers":{"Content-Type":"application/json"},"props":{}}' -u guest:guest
curl -X POST -H "content-type:application/json" http://localhost:15672/api/exchanges/%2f/amq.default/publish -d '{"vhost":"/","name":"amq.default","properties":{"delivery_mode":1,"headers":{"Content-Type":"application/json"}},"routing_key":"ConsumeTest","delivery_mode":"1","payload":"{\"treeId\":\"tree\",\"nodeId\":\"node2\",\"parentId\":\"node1\",\"childIds\":[],\"nodeType\":\"\",\"timestamp\":1,\"appJSON\":{}}","payload_encoding":"string","headers":{"Content-Type":"application/json"},"props":{}}' -u guest:guest

echo "Message published to ConsumeTest queue"

# Wait for the message to be consumed, processed and published to the ProduceTest queue
sleep 5

# Get the message from the ProduceTest queue using management API
echo "Consuming message from ProduceTest queue..."
curl -X POST -H "content-type:application/json" http://localhost:15672/api/queues/%2f/ProduceTest/get -d '{"vhost":"/","name":"ProduceTest","truncate":"50000","ackmode":"ack_requeue_false","encoding":"auto","count":"1"}' -u guest:guest| jq -r '.[].payload' >> /tmp/output.txt
grep -q 'tree' /tmp/output.txt && grep -q 'node1' /tmp/output.txt && grep -q 'node2' /tmp/output.txt && grep -q '1' /tmp/output.txt && grep -q '2' /tmp/output.txt

echo "Message consumed and processed successfully"

# Tear down the environment
docker compose -f docker-compose-buildtest.yml down

echo "Test completed successfully"
