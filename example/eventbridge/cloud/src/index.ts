import * as cdk from '@aws-cdk/core'
import * as pure from 'aws-cdk-pure'
import * as events from '@aws-cdk/aws-events'
import * as sink from '@aws-cdk/aws-events-targets'
import * as lambda from '@aws-cdk/aws-lambda'
import * as scud from 'aws-scud'
import * as path from 'path'


// ----------------------------------------------------------------------------
//
// Config
//
// ----------------------------------------------------------------------------
const app = new cdk.App()
const config = {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION,
  }
}

const stack = new cdk.Stack(app, 'swarm-example-eventbridge', { ...config })

// ----------------------------------------------------------------------------
//
// Resources
//
// ----------------------------------------------------------------------------

const Bus = (): pure.IPure<events.EventBus> => {
  const fBus = (): events.EventBusProps => ({
    eventBusName: "swarm-test",
  })
  return pure.iaac(events.EventBus)(fBus)
}

const MessagePattern = (
  eventBus: events.EventBus,
  fn: lambda.IFunction,
): pure.IaaC<events.Rule> => {
  const fPattern = (): events.RuleProps => ({
    eventBus,
    eventPattern: {
      account: [cdk.Aws.ACCOUNT_ID],
    },
    targets: [ new sink.LambdaFunction(fn) ],
  })
  return  pure.iaac(events.Rule)(fPattern)
}

const Consumer = (
  bus: events.EventBus,
): pure.IPure<events.Rule> => {
  const fConsumer = (): lambda.FunctionProps =>
    scud.handler.Go({
      sourceCodePackage: path.join(__dirname, '../..'),
      sourceCodeLambda: 'recv',
      timeout: cdk.Duration.seconds(30),
    })

  return scud.aws.Lambda(fConsumer)
    .flatMap(f => MessagePattern(bus, f))
}

pure.join(stack,
  Consumer(
    pure.join(stack, Bus()),
  ),
)
