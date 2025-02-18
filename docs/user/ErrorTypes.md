# Error Types
The Server library defines a set of "Error Types" that are used to categorize the errors that occur during the execution of the server. The error types can then be used by Consumers to determine how to handle the error. The error types are defined in the `ErrorType` enum in the `Server.appErrors` package, they are as follows:

- <a name="InvalidError">InvalidError</a> - The error is due to an invalid input fo some kind.
- <a name="SendError">SendError</a> - The error is due to an error that occurred while sending the output onwards.
- <a name="FullError">FullError</a> - The application has declared that it is in a "full" state and cannot accept any more input.