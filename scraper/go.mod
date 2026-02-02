module github.com/hammo/influScope/scraper

go 1.25.2

replace github.com/hammo/influScope/pkg => ../pkg

require (
	github.com/brianvoe/gofakeit/v6 v6.28.0
	github.com/hammo/influScope/pkg v0.0.0-00010101000000-000000000000
	github.com/upfluence/amqp v0.0.1
)

require (
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/upfluence/errors v0.2.15 // indirect
	github.com/upfluence/pkg/v2 v2.0.1 // indirect
	github.com/upfluence/stats v0.1.9 // indirect
	golang.org/x/time v0.13.0 // indirect
)
