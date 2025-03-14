# Consumers User Manual
## Introduction
Within this project an event driven architecture is used. Each application will consume data from a source, process the data and then send the data to a destination.

There are multiple consumers that can be used to consume the data. The current consumers are:

- AMQPOneConsumer (AMQP1.0 protocol) Client (based off the golang library https://github.com/Azure/go-amqp)
- RabbitMQ (AMQP0.9.1 protocol) Consumer Client (based off the golang library https://github.com/rabbitmq/amqp091-go)

Each Consumer client has specific configuration that is required to be set in order for the client to work correctly. The configuration is set in a JSON file that is passed to a Application that will use the Consumer client.

## Consumers Configuration
As a reminder the configuration file for an App as a whole has the following structure:

```json
{
    "AppConfig": <configuration for the application>,
    "ProducersSetup": <configuration for the producers to send the output to>,
    "ConsumersSetup": <configuration for the consumers to receive the input from - currently only works for a single consumer>
}
```
The consumers specific part of this configuration sits under the `ConsumersSetup` key and is used to configure the consumers that the JQExtractor will receive the input from (**Note**: *Currently only one Consumer is allowed*). The configuration has the following structure:
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
The `Type` field and corresponding `ConsumerConfig` object are dependent on the type of consumer.

Below are the current consumers that can be used and their configuration options.

### AMQPOneConsumer
The `AMQPOneConsumer` is a consumer that uses the AMQP1.0 protocol to receive messages from a queue. The configuration has the following structure:
```json
{
    "Type": "AMQPOneConsumer",
    "ConsumerConfig": {
        "Connection": <string - the connection string to the AMQP broker>,
        "Queue": <string - the name of the queue to consume from>,
        "MaxConcurrentMessages": <int, optional - the maximum number of messages to process concurrently, default is 1>,
        "OnValue": <boolean, optional - if true the consumer will grab data from AMQP1.0 packet "Value" field, if false the consumer will grab data from AMQP1.0 packet "Data" field, default is false>,
        "tlsConfig": <object, optional - the TLS configuration for the consumer>,
    }
}
```
See [TLS Configuration](./TLS_Config.md) for more information on the `tlsConfig` object.
#### Failure Handling
The `AMQPOneConsumer` will handle the following error types:

- `InvalidError` - The error is due to an invalid input of some kind. The AMQPOneConsumer will "Reject" the message and it will be sent to the dead letter queue.
- `SendError` - The error is due to an error that occurred while sending the output onwards. The AMQPOneConsumer will "Release" the message and it will be sent back to the queue with the delivery count incremented.
- `FullError` - The application has declared that it is in a "full" state and cannot accept any more input. The AMQPOneConsumer will "Modify" the message and it will be sent back to the queue with the delivery count incremented.

### RabbitMQConsumer
The `RabbitMQConsumer` is a consumer that uses the AMQP0.9.1 protocol to receive messages from a queue. The configuration has the following structure:
```json
{
    "Type": "RabbitMQConsumer",
    "ConsumerConfig": {
        "Connection": <string - the connection string to the RabbitMQ broker>,
        "Queue": <string - the name of the queue to consume from>,
    }
}
```

## Example Configuration

An example configuration file is as follows:
```json
{
    "AppConfig": <configuration for the application>,
    "ProducersSetup": <configuration for the producers to send the output to>,
    "ConsumersSetup": {
        "ConsumerConfigs": [
            {
                "Type": "AMQPOneConsumer",
                "ConsumerConfig": {
                    "Connection": "amqp://localhost:5672",
                    "Queue": "input_queue",
                    "MaxConcurrentMessages": 1,
                    "OnValue": false
                }
            }
        ]
    }
}
```

