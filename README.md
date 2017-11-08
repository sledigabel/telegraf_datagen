# telegraf_datagen

Data generator for Telegraf socket_listener

## State

Initial commit.
Doesn't do much.

## Usage

1. Setup Telegraf with the socket_listener input plugin
```
[[inputs.socket_listener]]
  service_address = "tcp://:8094"
  data_format = "influx"
```

2. Build

```
go build telegraf.go
```

3. Use

Doesn't do much for now apart from printing two little metrics
More will come soon.