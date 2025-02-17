# Group And Verify User Manual

## Introduction
The Group And Verify tool is an application that groups  tree like linked logs i.e. OTel Traces (https://opentelemetry.io/docs/concepts/signals/traces/). The application groups them together based on their "treeId" field and verifies that a complete tree has been collected using bi-directional links. The incoming data is expected to be singular JSON objects. The outgoing data will be in the form of a JSON array of JSON objects. The tool is designed to be run as a standalone application or as a part of a pipeline.

The workflow for the tool is shown in the sequence diagram below:
![Group And Verify Sequence Diagram](/docs/user/GroupAndVerify/GroupAndVerify_WorkFlow.svg)

The Group and Verify tool's current usage is to aid in transforming OTel Traces into a format that can be used by the [Protocol Verifier](https://github.com/xtuml/munin) i.e. Protocol Verifier audit event sequences. The tool will group the OTel Traces into trees and verify that the trees are complete. The trees will then be sent on for sequencing.

## Input Data Format
The incoming data will be required have the following format:
```json
{
    "treeId": <string>,
    "nodeId": <string>,
    "parentNodeId": <string>,
    "childIds": [<string>, ...],
    "nodeType": <string>, // optional, if the node type is in the parentVerifySet then the parent node is not expected to have a forward link to the child node
    "timestamp": <int>, //optional
    "appJSON": {
        <arbitrary fields>
    } 
}
```
For the example of OTel Traces:
- `treeId` is the `trace_id` of the trace
- `nodeId` is the `span_id` of the span
- `parentNodeId` is the `parent_span_id` of the span
- `timestamp` is the `end_time_unix_nano` of the span

The `appJSON` is arbitrary JSON object that holds the data that the user wants to be passed through the system (this can be updated by the system). The sequencer will add a specified (in config) field to the `appJSON` object that will hold the reference to the previous node in the sequence. For the example of Protocol Verifier audit event sequences this field will be `previousEventIds` and will hold the `eventId` of the previous event in the sequence within an array. An example of the input `appJSON` object for the Protocol Verifier is:
```json
{
    "jobName": "<string>",
    "jobId": "<UUID>",
    "eventId": "<UUID>",
    "eventType": "<string>",
    "timestamp": "<timestamp ISO string>",
    "applicationName": "<string>",
    <arbitrary non-conflicting fields>
}
```

## Output Data Format
The outgoing data will be in the form of a JSON array of JSON objects. The JSON objects will have the same format as the incoming data. The JSON array will contain all the nodes that are part of a tree. The trees will be sent to the configured producers i.e. on for sequencing.

```json
[
    {
        "treeId": <string>,
        "nodeId": <string>,
        "parentNodeId": <string>,
        "childIds": [<string>, ...],
        "nodeType": <string>, // optional, if the node type is in the parentVerifySet then the parent node is not expected to have a forward link to the child node
        "timestamp": <int>, //optional
        "appJSON": {
            <arbitrary fields>
        } 
    },
    ...
]
```

## Group and Verify Logic
The logic of the group and verify tool is split into two main parts:
- grouping the data from received nodes
- verifying the grouped data

### Grouping Logic
The grouping logic will follow the following steps:
- A node is received
- The "treeId" is identified
- If the "treeId" has not been seen before then a queue is created for that "treeId"
- The node is added to the unique "treeId" queue

The diagram below shows the grouping logic:
![Grouping Logic Diagram](/docs/user/GroupAndVerify/logic_grouping.svg)

### Verifying Logic
The basic premise of verification is that a tree is considered to be verified if all nodes apart from the root node have a bi-directional link to their parent node. This means that if a node, with "nodeId", has identified a parent node using "parentNodeId" then the parent node must have a "childId" in the "childIds" array that matches the "nodeId" of the child node.

The verification logic will follow the following steps for a single tree:
- A queue is created for the "treeId"
- A node is waiting in the queue and is received
- A timer is started for the tree
- Wait for another node
    - if the timer reaches a timeout value then notify that the tree is unverified and break
    - if a node is receivedcheck if all bi-directional links exist now
        - if all bi-directional links exist confirm the tree is verified and break
        - if not then continue waiting for nodes
- Output all currently collected nodes in the tree

The diagram below shows the verification logic:
![Verification Logic Diagram](/docs/user/GroupAndVerify/logic_verification.svg)

## Installation
### Building from Source
Prerequisites:
- Docker
- Internet connection

The docker compose file for the application (`GroupAndVerify/deploy/docker-compose.yml` - relative to the root of the repository) can be used to build the application. The application can be built using the following sequence of commands:
```bash
cd GroupAndVerify/deploy
docker compose build
```

### GitHub Container Registry
Prerequisites:
- Docker
- Internet connection

The Group And Verify tool is available as a Docker image. The image is available on the GitHub Container Registry (TODO: Add link). This image can be pulled using the following command:
- `docker pull ghcr.io/xtuml/groupandverify:latest`

## Configuration
A configuration file must be provided to the Group And Verify tool for it to run. The configuration file is a JSON file that has the following structure:
```json
{
    "AppConfig": <configuration for the application>,
    "ProducersSetup": <configuration for the producers to send the output to>,
    "ConsumersSetup": <configuration for the consumers to receive the input from - currently only works for a single consumer>
}
```

### AppConfig
The `AppConfig` section of the configuration file is used to configure the Group And Verify application. The configuration has the following structure:
```json
{
    "Timeout": <int, optional - the time in seconds to wait for a tree to be verifier, default is 2>,
    "MaxTrees": <int, optional - the maximum number of trees to be stored in memory and must be 0 or greater. If 0 then there is no maximum amount of trees, default is 0 i.e. no maximum>,
    "parentVerifySet": <array of objects, optional - the node types that are expected to have a parent links and the parent links are not expected to have a forward link to the child node. Specifies the expected number of children, default is []>,
}
```

The `Timeout` field is used to set the time in seconds to wait for a tree to be verified. If this is breached the tree will be sent to the producer as is and the application that sent the data will be notified.

The `MaxTrees` field is used to set the maximum number of trees that the application will process at one time, if this limit is reached the application that sent the data will be notified.

The `parentVerifySet` field is used to set the node types that are expected to have a parent links and the parent links are not expected to have a forward links to the child nodes. Specifies the expected number of children mapped to nodeType. The structure of the object in the array is

```json
{
    "nodeType": <string>,
    "expecteddChildren": <int>
}
```

An example configuration for AppConfig is as follows:
```json
{
    "AppConfig": {
        "Timeout": 2,
        "MaxTrees": 0,
        "parentVerifySet": [
            {
                "nodeType": "nodeType1", "expecteddChildren": 1
            },
            {
                "nodeType":"nodeType2",
                "expecteddChildren": 2
            }
        ]
    },
}
```

### ProducersSetup
The `ProducersSetup` section of the configuration file is used to configure the producer that the Group And Verify will send the output to. The configuration has the following structure:
```json
{
    "IsMapping": <boolean - not used by the Group And Verify>,
    "ProducerConfigs": <array of objects that contain the fields: Type, ProducerConfig>,
}
```

The objects in the `ProducerConfigs` array have the following structure:
```json
{
    "Type": <string - the type of producer to use. From a set list of types>,
    "ProducerConfig": <object - the configuration for the producer>,
}
```

The `Type` field and corresponding `ProducerConfig` object are dependent on the type of producer. These can be found in the [Producers Documentation](/docs/user/Producers_User_Manual.md).

### ConsumersSetup
The `ConsumersSetup` section of the configuration file is used to configure the consumers that the Group And Verify will receive the input from. The configuration has the following structure:
```json
{
    "ConsumerConfigs": <array of objects that contain the fields: Type, ConsumerConfig>,
}
```

The objects in the `ConsumerConfigs` array have the following structure:
```json
{
    "Type": <string - the type of consumer to use. From a set list of types>,
    "ConsumerConfig": <object - the configuration for the consumer>,
}
```

The `Type` field and corresponding `ConsumerConfig` object are dependent on the type of consumer. These can be found in the [Consumer Documentation](/docs/user/Consumers_User_Manual.md).

### Example Configuration

An example configuration file is as follows:
```json
{
    "AppConfig": {
        "Timeout": 2,
        "MaxTrees": 0,
        "parentVerifySet": [
            "nodeType1",
            "nodeType2"
        ]
    },
    "ProducersSetup": {
        "IsMapping": false,
        "ProducerConfigs": [
            {
                "Type": "AMQP1.0",
                "ProducerConfig": {
                    "Address": "amqp://localhost:5672",
                    "Queue": "output_queue"
                }
            }
        ]
    },
    "ConsumersSetup": {
        "ConsumerConfigs": [
            {
                "Type": "AMQP1.0",
                "ConsumerConfig": {
                    "Address": "amqp://localhost:5672",
                    "Queue": "input_queue"
                }
            }
        ]
    }
}
```

## Running the Application
The Group And Verify application can be run using the docker compose file found at `GroupAndVerify/deploy/docker-compose.yml` (relative to the root of the repository). All files required for the application to run must be in the `GroupAndVerify/deploy/config` directory. The application can be run using the following command (starting from the root directory of the repository):
```bash
cd GroupAndVerify/deploy
docker compose up -d
```