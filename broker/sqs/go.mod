module github.com/fogfish/swarm/broker/sqs

go 1.22

replace github.com/fogfish/swarm => ../../

replace github.com/fogfish/swarm/queue => ../../queue

replace github.com/fogfish/swarm/qtest => ../../qtest

require (
	github.com/aws/aws-sdk-go-v2 v1.31.0
	github.com/aws/aws-sdk-go-v2/config v1.27.37
	github.com/aws/aws-sdk-go-v2/service/sqs v1.35.1
	github.com/fogfish/swarm v0.0.0-00010101000000-000000000000
	github.com/fogfish/swarm/qtest v0.0.0-00010101000000-000000000000
	github.com/fogfish/swarm/queue v0.0.0-00010101000000-000000000000
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.17.35 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.23.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.27.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.31.1 // indirect
	github.com/aws/smithy-go v1.21.0 // indirect
	github.com/fogfish/curie v1.8.2 // indirect
	github.com/fogfish/faults v0.2.0 // indirect
	github.com/fogfish/golem/hseq v1.2.0 // indirect
	github.com/fogfish/golem/optics v0.13.0 // indirect
	github.com/fogfish/golem/pure v0.10.1 // indirect
	github.com/fogfish/guid/v2 v2.0.4 // indirect
	github.com/fogfish/it/v2 v2.0.2 // indirect
)
