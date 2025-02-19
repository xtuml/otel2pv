# Logging
The applications available in this project have standard logging capabilities. The logging is done using the slog library. The logging is done at different levels, such as `DEBUG`, `INFO`, `WARN`, `ERROR` (in ascending order of severity). The logging level can be set using the `LOG_LEVEL` environment variable. The logging level can be set to any of the following values:
- `DEBUG`
- `INFO`
- `WARN`
- `ERROR`

The default logging level is `DEBUG`. The logs are printed to the standard output. The logs are printed in the following format:
```json
{
    "time": <ISO 8601 formatted time>,
    "level": <log level>,
    "msg": <log message>,
    "logger_info": {"logger_name"<logger name>},
    "details": <additional details>
}
```

The logs can also be written to a file. The file path can be set using the `LOG_FILE` environment variable. If the `LOG_FILE` environment variable is not set, the logs are printed to the standard output.