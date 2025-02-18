# Producers User Manual
## Introduction
Within this project an event driven architecture is used. Each application will consume data from a source, process the data and then send the data to a destination.

There are multiple producers that can be used to produce the data. The current producers are:

- AMQPOneProducer (AMQP1.0 protocol) Client (based off the golang library https://github.com/Azure/go-amqp)
- RabbitMQ (AMQP0.9.1 protocol) Producer Client (based off the golang library https://github.com/rabbitmq/amqp091-go)
- HTTP Producer Client (using the standard golang http library)

Each Producer client has specific configuration that is required to be set in order for the client to work correctly. The configuration is set in a JSON file that is passed to a Application that will use the Producer client.

## Producers Configuration
As a reminder the configuration file for an App as a whole has the following structure:

```json
{
    "AppConfig": <configuration for the application>,
    "ProducersSetup": <configuration for the producers to send the output to>,
    "ConsumersSetup": <configuration for the consumers to receive the input from - currently only works for a single consumer>
}
```

The producers specific part of this configuration sits under the `ProducersSetup` key and is used to configure the producers that an application will send the output to. The configuration has the following structure:
```json
{
    "IsMapping": <boolean, optional - if true a key/identifier can be used by an Application to map to a specific producer, if false the producers will be used in a round robin fashion. default is false>,
    "ProducerConfigs": <array of objects that contain the fields: Type, ProducerConfig, and optional Map>,
}
```

The objects in the `ProducerConfigs` array have the following structure:
```json
{
    "Type": <string - the type of producer to use. From a set list of types>,
    "ProducerConfig": <object - the configuration for the producer>,
    "Map": <string, optional - the key/identifier to map to this producer>
}
```
The `Type` field and corresponding `ProducerConfig` object are dependent on the type of producer. The `Map` field is required if `IsMapping` is set to true and will be ignored if `IsMapping` is set to false.

Below are the current producers that can be used and their configuration options.

### AMQPOneProducer

The `AMQPOneProducer` is a producer that uses the AMQP1.0 protocol to send messages to a queue. The configuration has the following structure:
```json
{
    "Type": "AMQPOneProducer",
    "ProducerConfig": {
        "Connection": <string - the connection string to the AMQP broker>,
        "Queue": <string - the name of the queue to send to>,
        "MessageHeaders": <object, optional - the headers to add to the message>
    },
    "Map": <string, optional - the key/identifier to map to this producer>
}
```

The `MessageHeaders` object is used to add headers to the message. The object should have the following structure:
```json
{
    "Durable": <bool, optional - if true the message will be saved to disk by the broker, default is true>,
    "Priority": <int, optional - the AMQP1.0 protocol priority of the message, default is 4>,
    "TTL": <int, optional - the time in milliseconds that the message will be stored by the broker, default is unlimited>,
    "FirstAcquirer": <bool, optional - if true the message has not been acquired by any other consumer, default is false>,
}


### RabbitMQProducer

The `RabbitMQProducer` is a producer that uses the AMQP0.9.1 protocol to send messages to a queue. The configuration has the following structure:
```json
{
    "Type": "RabbitMQProducer",
    "ProducerConfig": {
        "Connection": <string - the connection string to the RabbitMQ broker>,
        "Queue": <string - the name of the queue to send to>,
    },
    "Map": <string, optional - the key/identifier to map to this producer>
}
```

### HTTPProducer

The `HTTPProducer` is a producer that uses the standard golang http library to send messages to a HTTP endpoint. The configuration has the following structure:
```json
{
    "Type": "HTTPProducer",
    "ProducerConfig": {
        "URL": <string - the URL to send the data to>,
        "numRetries": <int, optional - the number of times to retry sending the data, default is 3>,
        "timeout": <int, optional - the timeout in seconds for the request, default is 10>,
        "initialRetryInterval": <float, optional - the initial retry interval in seconds (for exponential back-off), default is 1>,
	    "retryIntervalmultiplier" <float, optional - the multiplier for the retry interval (for exponential back-off), default is 1>
    },
    "Map": <string, optional - the key/identifier to map to this producer>
}
```
The HTTPProducer can perform configured expontential back-off retries if the request fails. The `initialRetryInterval` and `retryIntervalmultiplier` fields are used to configure the back-off. The back-off is calculated as `initialRetryInterval * [0.5,1.5]retryIntervalmultiplier^retryNumber`. The intial retry interval is multiplied by a randomisation factor between 0.5 and 1.5 to prevent all clients from retrying at the same time. The retry interval is then multiplied by the retryIntervalmultiplier to the power of the retry number to increase the interval between retries.

### Example Configuration

An example configuration file is as follows:
```json
{
    "AppConfig": <configuration for the application>,
    "ConsumersSetup": <configuration for the consumers>,
    "ProducersSetup": {
        "IsMapping": false,
        "ProducerConfigs": [
            {
                "Type": "AMQP1.0",
                "ProducerConfig": {
                    "Address": "amqp://localhost:5672",
                    "Queue": "output_queue"
                }
            },
            {
                "Type": "RabbitMQ",
                "ProducerConfig": {
                    "Address": "amqp://localhost:5672",
                    "Queue": "output_queue"
                }
            },
            {
                "Type": "HTTP",
                "ProducerConfig": {
                    "URL": "http://localhost:8080",
                    "numRetries": 3,
                    "timeout": 10
                }
            }
        ]
    }
}
```




