package eventddb

import (
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/fogfish/scud"
)

//------------------------------------------------------------------------------
//
// AWS CDK Sink Construct
//
//------------------------------------------------------------------------------

type Sink struct {
	constructs.Construct
	Handler awslambda.IFunction
}

type SinkProps struct {
	Table       awsdynamodb.ITable
	Lambda      *scud.FunctionGoProps
	EventSource *awslambdaeventsources.DynamoEventSourceProps
}

func NewSink(scope constructs.Construct, id *string, props *SinkProps) *Sink {
	sink := &Sink{Construct: constructs.NewConstruct(scope, id)}

	sink.Handler = scud.NewFunctionGo(sink.Construct, jsii.String("Func"), props.Lambda)

	eventsource := &awslambdaeventsources.DynamoEventSourceProps{
		StartingPosition: awslambda.StartingPosition_LATEST,
	}
	if props.EventSource != nil {
		eventsource = props.EventSource
	}

	source := awslambdaeventsources.NewDynamoEventSource(props.Table, eventsource)

	sink.Handler.AddEventSource(source)

	return sink
}

//------------------------------------------------------------------------------
//
// AWS CDK Stack Construct
//
//------------------------------------------------------------------------------

type ServerlessStackProps struct {
	*awscdk.StackProps
	Version string
	System  string
}

type ServerlessStack struct {
	awscdk.Stack
	acc           int
	removalPolicy awscdk.RemovalPolicy
	Table         awsdynamodb.ITable
}

func NewServerlessStack(app awscdk.App, id *string, props *ServerlessStackProps) *ServerlessStack {
	sid := *id
	if props.Version != "" {
		sid = sid + "-" + props.Version
	}

	stack := &ServerlessStack{
		Stack:         awscdk.NewStack(app, jsii.String(sid), props.StackProps),
		removalPolicy: awscdk.RemovalPolicy_RETAIN,
	}

	if strings.HasPrefix(props.Version, "pr") {
		stack.removalPolicy = awscdk.RemovalPolicy_DESTROY
	}

	return stack
}

func (stack *ServerlessStack) NewTable(tableName ...string) awsdynamodb.ITable {
	name := awscdk.Aws_STACK_NAME()
	if len(tableName) > 0 {
		name = &tableName[0]
	}

	stack.Table = awsdynamodb.NewTable(stack.Stack, jsii.String("Table"),
		&awsdynamodb.TableProps{
			TableName: name,
			PartitionKey: &awsdynamodb.Attribute{
				Type: awsdynamodb.AttributeType_STRING,
				Name: jsii.String("prefix"),
			},
			SortKey: &awsdynamodb.Attribute{
				Type: awsdynamodb.AttributeType_STRING,
				Name: jsii.String("suffix"),
			},
			BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
			RemovalPolicy: stack.removalPolicy,
			Stream:        awsdynamodb.StreamViewType_NEW_IMAGE,
		},
	)

	return stack.Table
}

func (stack *ServerlessStack) AddTable(tableName string) awsdynamodb.ITable {
	stack.Table = awsdynamodb.Table_FromTableName(stack.Stack, jsii.String("Table"),
		jsii.String(tableName),
	)

	return stack.Table
}

func (stack *ServerlessStack) NewGlobalTable(tableName ...string) awsdynamodb.ITable {
	name := awscdk.Aws_STACK_NAME()
	if len(tableName) > 0 {
		name = &tableName[0]
	}

	stack.Table = awsdynamodb.NewTableV2(stack.Stack, jsii.String("Table"),
		&awsdynamodb.TablePropsV2{
			TableName: name,
			PartitionKey: &awsdynamodb.Attribute{
				Type: awsdynamodb.AttributeType_STRING,
				Name: jsii.String("prefix"),
			},
			SortKey: &awsdynamodb.Attribute{
				Type: awsdynamodb.AttributeType_STRING,
				Name: jsii.String("suffix"),
			},
			Billing:       awsdynamodb.Billing_OnDemand(),
			RemovalPolicy: stack.removalPolicy,
			DynamoStream:  awsdynamodb.StreamViewType_NEW_IMAGE,
		},
	)

	return stack.Table
}

func (stack *ServerlessStack) AddGlobalTable(tableName string) awsdynamodb.ITable {
	stack.Table = awsdynamodb.TableV2_FromTableName(stack.Stack, jsii.String("Table"),
		jsii.String(tableName),
	)

	return stack.Table
}

func (stack *ServerlessStack) NewSink(props *SinkProps) *Sink {
	if stack.Table == nil {
		panic("Table is not defined.")
	}

	props.Table = stack.Table

	stack.acc++
	name := "Sink" + strconv.Itoa(stack.acc)
	sink := NewSink(stack.Stack, jsii.String(name), props)

	return sink
}

//------------------------------------------------------------------------------
//
// AWS CDK App Construct
//
//------------------------------------------------------------------------------

type ServerlessApp struct {
	awscdk.App
}

func NewServerlessApp() *ServerlessApp {
	app := awscdk.NewApp(nil)
	return &ServerlessApp{App: app}
}

func (app *ServerlessApp) NewStack(name string, props ...*awscdk.StackProps) *ServerlessStack {
	config := &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
			Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
		},
	}

	if len(props) == 1 {
		config = props[0]
	}

	return NewServerlessStack(app.App, jsii.String(name), &ServerlessStackProps{
		StackProps: config,
		Version:    FromContextVsn(app),
		System:     name,
	})
}

func FromContext(app awscdk.App, key string) string {
	val := app.Node().TryGetContext(jsii.String(key))
	switch v := val.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func FromContextVsn(app awscdk.App) string {
	vsn := FromContext(app, "vsn")
	if vsn == "" {
		return "latest"
	}

	return vsn
}
