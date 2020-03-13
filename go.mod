module github.com/owncloud/ocis-proxy

go 1.13

require (
	contrib.go.opencensus.io/exporter/jaeger v0.2.0
	contrib.go.opencensus.io/exporter/ocagent v0.6.0
	contrib.go.opencensus.io/exporter/zipkin v0.1.1
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/golang/protobuf v1.3.4 // indirect
	github.com/micro/cli/v2 v2.1.2
	github.com/micro/go-micro/v2 v2.2.0
	github.com/miekg/dns v1.1.28 // indirect
	github.com/oklog/run v1.1.0
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/owncloud/ocis-accounts v0.0.0-20200309101757-1c32865f6446
	github.com/owncloud/ocis-konnectd v0.0.0-20200303180152-937016f63393
	github.com/owncloud/ocis-pkg/v2 v2.0.2
	github.com/prometheus/client_golang v1.2.1
	github.com/restic/calens v0.2.0
	github.com/spf13/viper v1.6.2
	go.opencensus.io v0.22.2
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073 // indirect
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
)
