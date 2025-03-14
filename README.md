# CDS-OTel-To-PV

## Introduction
The purpose of this project is to convert Open Telemetry (OTel) traces (https://opentelemetry.io/docs/concepts/signals/traces/) and their underlying spans (https://opentelemetry.io/docs/concepts/signals/traces/#spans) into sequenced data for the [Protocol Verifier](https://github.com/xtuml/munin) (PV). The system processes logged call trees presented in JSON format, making it versatile and generalizable to any similar data structure. By leveraging a series of applications, the system ensures that the data is extracted, grouped, verified (a timeout may occur), and sequenced before being sent to the Protocol Verifier.

The system is designed to be modular, allowing for the addition of new applications and the modification of existing ones. This flexibility ensures that the system can be adapted to various use cases involving logged call trees. Currently the system consists of three applications: [JQExtractor](#jqextractor), [Group and Verify](#group-and-verify), and [Sequencer](#sequencer).

## Server Library
The [Server Library](/Server) provides the basis for the application above. It is a library that provides the functionality to create a server that can receive data from a producer and send data to a consumer. The library is designed to be used in a modular way, allowing for the creation of a server that can be used to send data between multiple applications. 

The library relies on the high level concept of a **Consumer** receiving messages from a source. The messages are then sent on to  a **PipeServer** that can do any amount of processing on that message. Once the processing is completed the processed message is then sent on to a **Producer** that sends the data to a destination. The library is designed to be used in a modular way, allowing for the creation of a server that can be used to send data between multiple applications.

![Server Library Overview](/docs/README/server_library_arch.svg)

## Applications in the System

### JQExtractor
The JQExtractor application extracts JSON objects from streamed JSON data based on multiple JQ queries. The queries are written in the JQ language (https://jqlang.org/manual/) but specifically use the [gojq](https://github.com/itchyny/gojq) package. The extracted JSON objects are then sent to configured destinations, such as an AMQP1.0 queue or an HTTP endpoint. Further details can be found in the [JQExtractor User Manual](/docs/user/JQExtractor/JQExtractor_User_Manual.md).

#### Workflow
1. Receives JSON packets.
2. Applies JQ queries to extract relevant information.
3. Sends the extracted JSON packets to destinations defined by specific identifiers.



### Group and Verify
The Group and Verify application groups tree-like linked logs (e.g., OTel Traces with Spans) based on their `tree_id` field (`trace_id` for OTel) and verifies that a complete tree has been collected using bi-directional links i.e. the parent references the child (`childNodeIds` - child `span_id`'s as applied to OTel) and the child references the parent (`parentNodeId` - `parent_span_id` as applied to OTel). This verification process allows for faster processing as it can then be confirmed that the tree is complete and can be sent on to be sequenced. Further details can be found in the [Group and Verify User Manual](/docs/user/GroupAndVerify/GroupAndVerify_User_Manual.md).

#### Workflow
1. Receives singular JSON packets.
2. Groups packets by "tree_id".
3. Verifies the completeness of the groups and the existence of bi-directional links.
4. Sends grouped JSON arrays to the next destination.
5. Notifies in logs the verification status.

### Sequencer
The Sequencer application sequences tree-like linked logs (e.g., OTel Traces and Spans) into their causal sequences (Protocol Verifier Audit Event sequences - see [Audit Event Tutorial](https://github.com/xtuml/plus2json/blob/main/doc/tutorial/AuditEventTopologyTutorial.pdf)) i.e. which log of an event when complete caused the next log to occur. The sequencer ensures that the data is sequenced in a straight line sequence, with each node having a single parent and a single child, unless a starting node. The sequenced data is then sent to the Protocol Verifier in a format that can be ingested by the PV. Further details can be found in the [Sequencer User Manual](/docs/user/Sequencer/Sequencer_User_Manual.md).

#### Workflow
1. Receives grouped JSON arrays.
2. Sequences the arrays.
3. Sends the sequenced arrays to the Protocol Verifier.
4. Notifies the source of successful processing.

## System Integration
The applications in the system currenlty communicate with each other using AMQP1.0 (however with development other protocols e.g. Kafka could also be used), ensuring reliable message passing and processing. The overall workflow is as follows:

1. **JQExtractor** receives and processes JSON packets, extracting relevant data and sending it to the **Group and Verify** application.
2. **Group and Verify** groups and verifies the data, ensuring completeness and correctness (unless timeout occurs), then sends the grouped data to the **Sequencer**.
3. **Sequencer** sequences the data and sends it to the Protocol Verifier, completing the conversion process.

This modular approach allows for flexibility and scalability, making it possible to adapt the system to various requirements.

![System Overview](/docs/README/otel_to_pv_arch.svg)