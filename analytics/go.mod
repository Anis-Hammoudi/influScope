module github.com/hammo/influScope/analytics

go 1.24.0

replace github.com/hammo/influScope/gen/analytics => ../gen/analytics

replace github.com/hammo/influScope/pkg => ../pkg

require (
	github.com/hammo/influScope/gen/analytics v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.23.2
	google.golang.org/grpc v1.78.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
