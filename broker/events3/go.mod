module github.com/fogfish/swarm/broker/events3

go 1.22

replace github.com/fogfish/swarm => ../../

replace github.com/fogfish/swarm/queue => ../../queue

replace github.com/fogfish/swarm/qtest => ../../qtest

require (
	github.com/aws/aws-cdk-go/awscdk/v2 v2.160.0
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/constructs-go/constructs/v10 v10.3.0
	github.com/aws/jsii-runtime-go v1.103.1
	github.com/fogfish/guid/v2 v2.0.4
	github.com/fogfish/scud v0.10.1
	github.com/fogfish/swarm v0.0.0-00010101000000-000000000000
	github.com/fogfish/swarm/queue v0.0.0-00010101000000-000000000000
)

require (
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/cdklabs/awscdk-asset-awscli-go/awscliv1/v2 v2.2.202 // indirect
	github.com/cdklabs/awscdk-asset-kubectl-go/kubectlv20/v2 v2.1.2 // indirect
	github.com/cdklabs/awscdk-asset-node-proxy-agent-go/nodeproxyagentv6/v2 v2.1.0 // indirect
	github.com/cdklabs/cloud-assembly-schema-go/awscdkcloudassemblyschema/v38 v38.0.1 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/fogfish/curie v1.8.2 // indirect
	github.com/fogfish/faults v0.2.0 // indirect
	github.com/fogfish/golem/hseq v1.2.0 // indirect
	github.com/fogfish/golem/optics v0.13.0 // indirect
	github.com/fogfish/golem/pure v0.10.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/yuin/goldmark v1.5.3 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.20.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.23.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
)
