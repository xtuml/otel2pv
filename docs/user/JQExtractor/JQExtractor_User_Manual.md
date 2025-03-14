# JQExtractor User Manual
## Introduction
JQExtractor is a tool that extracts JSON objects from streamed JSON objects based on multiple JQ queries. The query is written in the JQ language (https://jqlang.org/manual/). JQ is a lightweight and flexible command-line JSON processor. JQExtractor uses a Go JQ library (https://github.com/itchyny/gojq) as the basis of the extraction of the JSON objects from a JSON file. The output of each query is sent to a configured location (e.g. an AMQP1.0 queue). The tool is designed to be run as a standalone application or as a part of a pipeline.

The workflow for two queries (the basic premise can be extended to more than two queries) is shown in the sequence diagram below:
![JQExtractor Sequence Diagram](/docs/user/JQExtractor/JQExtractor_WorkFlow.svg)

The JQExtractor will process all queries and send all extracted packets on to the configured destinations and then send confirmation to the source that the sent JSON packet has been processed.
## Installation
### Building from Source
Prerequisites:
- Docker
- Internet connection

The docker compose file for the application (`JQExtractor/deploy/docker-compose.yml` - relative to the rooot of the repository) can be used to build the application. The application can be built using the following sequence of commands:
```bash
cd JQExtractor/deploy
docker compose build
```
### GitHub Container Registry
Prerequisites:
- Docker
- Internet connection

The JQExtractor tool is available as a Docker image. The image is available on the GitHub Container Registry (TODO: Add link). This image can be pulled using the following command:
- `docker pull ghcr.io/xtuml/jqextractor:latest`
## Configuration
A configuration file must be provided to the JQExtractor for it to run. The configuration file is a JSON file that has the following structure:
```json
{
    "AppConfig": <configuration for the application>,
    "ProducersSetup": <configuration for the producers to send the output to>,
    "ConsumersSetup": <configuration for the consumers to receive the input from - currently only works for a single consumer>
}
```
### AppConfig
The `AppConfig` section of the configuration file is used to configure the JQExtractor application. The configuration has the following structure:
```json
{
    "JQQueryStrings": <an object that maps a key/identifier to an object with information about the JQ query, the input format and a schema to validate the output>,
}
```
The key/identifier is used to map to a Producer (configured in the `ProducersSetup` section) to send the output to.

The object for each JQ query has the following structure:
```json
{
    "jq": "<string - the JQ query or a file path to the JQ query file>",
    "type": "<string, optional - takes the values 'file' for providing a file path or 'string' for providing the JQ query as a string. default is 'string'>",
    "validate": "<string, optional - filepath to a JSON schema file to validate the output of the JQ query>",
}
```
Example:
```json
{
    "AppConfig": {
        "JQQueryStrings": {
            "query1": {
                "jq": ".",
                "validate": "schema.json"
            },
            "query2": {
                "jq": "jq_query.jq",
                "type": "file"
            }
        }
    }
}
```
### ProducersSetup
The `ProducersSetup` section of the configuration file is used to configure the producers that the JQExtractor will send the output to. The configuration has the following structure:
```json
{
    "IsMapping": <boolean, optional - if true the key/identifier in the JQQueryStrings will be used to map to the producer, if false the producers will be used in a round robin fashion. default is false>,
    "ProducerConfigs": <array of objects that contain the fields: Type, ProducerConfig, and optional Map>,
}
```
The objects in the `ProducerConfigs` array have the following structure:
```json
{
    "Type": <string - the type of producer to use. From a set list of types>,
    "ProducerConfig": <object - the configuration for the producer>,
    "Map": <string, optional - the key/identifier in the JQQueryStrings to map to this producer>
}
```
The `Map` field is required if `IsMapping` is set to true and will be ignored if `IsMapping` is set to false.

The `Type` field and corresponding `ProducerConfig` object are dependent on the type of producer. These can be found in the [Prodcuers Documentation](/docs/user/Producers_User_Manual.md).

### ConsumersSetup
The `ConsumersSetup` section of the configuration file is used to configure the consumer that the JQExtractor will receive the input from. The configuration has the following structure:
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
        "JQQueryStrings": {
            "query1": {
                "jq": ".",
                "validate": "schema.json"
            },
            "query2": {
                "jq": "jq_query.jq",
                "type": "file"
            }
        }
    },
    "ProducersSetup": {
        "IsMapping": false,
        "ProducerConfigs": [
            {
                "Type": "AMQP1.0",
                "ProducerConfig": {
                    "Address": "amqps://localhost:5672",
                    "Queue": "output_queue",
                    "tlsConfig": {
                        "certFile": "/config/client.crt",
                        "keyFile": "/config/client.key",
                        "clientCaFile": "/config/ca.crt",
                        "rootCaFile": "/config/root.crt",
                        "tlsVersion": "TLSv1.2"
                    }
                }
            }
        ]
    },
    "ConsumersSetup": {
        "ConsumerConfigs": [
            {
                "Type": "AMQP1.0",
                "ConsumerConfig": {
                    "Address": "amqps://localhost:5672",
                    "Queue": "input_queue",
                    "tlsConfig": {
                        "certFile": "/config/client.crt",
                        "keyFile": "/config/client.key",
                        "clientCaFile": "/config/ca.crt",
                        "rootCaFile": "/config/root.crt",
                        "tlsVersion": "TLSv1.2"
                    }
                }
            }
        ]
    }
}
```

The config file must be provided in the directory `JQExtractor/deploy/config` as `config.conf`. Any extra files such as .jq files or schema files refered to in the config must also be provided in the `JQExtractor/deploy/config` directory and their paths must be relative to the `config` directory of the running container i.e. `/config/<filename>`.

## Running the Application
The JQExtractor application can be run using the docker compose file found at `JQExtractor/deploy/docker-compose.yml` (relative to the root of the repository). All files required for the application to run must be in the `JQExtractor/deploy/config` directory. The application can be run using the following command (starting from the root directory of the repository):
```bash
cd JQExtractor/deploy
docker compose up -d
```

## Return To Sender Cases
The following outlines the cases in which incoming data will be returned to the sender:

1. The incoming data is not in the correct format i.e. the data is not valid JSON - in this case an [InvalidError](/docs/user/ErrorTypes.md#InvalidError) will be raised internally (see consumer specific documentation in [Consumers Manual](/docs/user/Consumers_User_Manual.md) for how this error is handled)
2. One of the jq strings definied in the config produces an error - in this case a [InvalidError](/docs/user/ErrorTypes.md#InvalidError) will be raised internally (see consumer specific documentation in [Consumers Manual](/docs/user/Consumers_User_Manual.md) for how this error is handled)
3. One of the jq strings definied in the config produces an output that does not match the input schema - in this case an [InvalidError](/docs/user/ErrorTypes.md#InvalidError) will be raised internally (see consumer specific documentation in [Consumers Manual](/docs/user/Consumers_User_Manual.md) for how this error is handled)
4. The chosen producer failed to send the message on to the next destination - in this case a [SendError](/docs/user/ErrorTypes.md#SendError) will be raised internally (see producer specific documentation in [Consumers Manual](/docs/user/Consumers_User_Manual.md) for how this error is handled)

## Logging
Logging can be configured as per the [Logging Documentation](/docs/user/Logging.md).
