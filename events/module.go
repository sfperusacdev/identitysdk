package events

import "go.uber.org/fx"

var PublisherModule = fx.Module(
	"events-publisher",
	fx.Provide(
		fx.Annotate(NewPublisher, fx.As(new(Publisher))),
	),
)

var ConsumerModule = fx.Module(
	"events-consumer",
	fx.Provide(
		fx.Annotate(mapConsumers, fx.ParamTags(ConsumerGroupTag)),
	),
	fx.Invoke(StartEventBusConsumers),
)

var Module = fx.Module(
	"events",
	PublisherModule,
	ConsumerModule,
)
