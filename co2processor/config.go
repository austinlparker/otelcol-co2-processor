package co2processor

import (
	"go.opentelemetry.io/collector/component"
)

type Config struct {
	filterconfig.MatchConfig `mapstructure:",squash"`
}
