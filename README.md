## About this project
This is a nostr relay with one database for each country. Therefore, when someone makes a connection the user's IP will be used to defined the user's country and only give events associated with that specific country.

For example, if your IP address is from Portugal then you will only receive events that are within Portugal's database.

## Running from the command line
```bash
export BASE_DOMAIN= && go build && ./countries
```

## Test this application
If you want to do some rough tests you can use the [nak](https://github.com/fiatjaf/nak) tool which allows you to perform WebSocket requests from the command line in a very easy and straightfoward way.

## Add an event with default parameters and pretty format response
```bash
nak event ws://localhost:40404 | jq
```

## Get relay's info and pretty format response
```bash
nak relay ws://localhost:40404 | jq
```

## Delete an event and pretty format response
```bash
nak event -k 5 -e <id> <relay> | jq
```
The *\<relay\>* of this project can be expressed as **ws://localhost:40404**
