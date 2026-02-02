module github.com/hammo/influScope/indexer

go 1.24.0

replace github.com/hammo/influScope/pkg => ../pkg

require (
	github.com/elastic/go-elasticsearch/v7 v7.17.10
	github.com/prometheus/client_golang v1.23.2
	github.com/upfluence/amqp v0.0.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/upfluence/errors v0.2.15 // indirect
	github.com/upfluence/pkg/v2 v2.0.1 // indirect
	github.com/upfluence/stats v0.1.9 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/time v0.13.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)
