module github.com/hammo/influScope/analytics

go 1.25.2

replace github.com/hammo/influScope/gen/analytics => ../gen/analytics

replace github.com/hammo/influScope/pkg => ../pkg

require (
	github.com/hammo/influScope/gen/analytics v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.78.0
)

require (
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
