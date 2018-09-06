# dkron-processor-agent

dkron processor plugin

## Installation

```sh
go get -u github.com/tengattack/dkron-processor-agent
sudo mkdir -p /etc/dkron/plugins
sudo cp ~/go/bin/dkron-processor-agent /etc/dkron/plugins
```

## Usage

```js
{
  // put in dkron job configuration
  // ...
  "processors": {
    "agent": {
      "forward": true,
      "dsn": "udp://logstash.server:10001"
    }
  },
  // ...
}
```

## License

MIT
