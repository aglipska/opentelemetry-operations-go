// Copyright 2022 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/otelcol/otelcoltest"

	"github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/collector"
)

func TestLoadConfig(t *testing.T) {
	gcpType, err := component.NewType("googlecloud")
	assert.NoError(t, err)
	factories, err := otelcoltest.NopFactories()
	assert.Nil(t, err)

	factory := newFactory(gcpType)
	factories.Exporters[gcpType] = factory
	cfg, err := otelcoltest.LoadConfigAndValidate(path.Join("..", "testdata", "config.yaml"), factories)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Exporters), 2)

	r0 := cfg.Exporters[component.NewID(gcpType)].(*testExporterConfig)
	defaultConfig := factory.CreateDefaultConfig().(*testExporterConfig)
	assert.Equal(t, sanitize(r0), sanitize(defaultConfig))

	r1 := cfg.Exporters[component.NewIDWithName(gcpType, "customname")].(*testExporterConfig)
	assert.Equal(t, sanitize(r1),
		&testExporterConfig{
			Config: collector.Config{
				ProjectID: "my-project",
				UserAgent: "opentelemetry-collector-contrib {{version}}",
				TraceConfig: collector.TraceConfig{
					ClientConfig: collector.ClientConfig{
						Endpoint:     "test-trace-endpoint",
						UseInsecure:  true,
						GRPCPoolSize: 1,
					},
				},
				MetricConfig: collector.MetricConfig{
					ClientConfig: collector.ClientConfig{
						Endpoint:     "test-metric-endpoint",
						UseInsecure:  true,
						GRPCPoolSize: 1,
					},
					Prefix:                     "prefix",
					SkipCreateMetricDescriptor: true,
					KnownDomains: []string{
						"googleapis.com", "kubernetes.io", "istio.io", "knative.dev",
					},
					InstrumentationLibraryLabels:     true,
					CreateMetricDescriptorBufferSize: 10,
					ServiceResourceLabels:            true,
					CumulativeNormalization:          true,
				},
				LogConfig: collector.LogConfig{
					ClientConfig: collector.ClientConfig{
						GRPCPoolSize: 1,
					},
					DefaultLogName:        "foo-log",
					ServiceResourceLabels: true,
				},
			},
		})
}

func sanitize(cfg *testExporterConfig) *testExporterConfig {
	cfg.Config.MetricConfig.MapMonitoredResource = nil
	cfg.Config.MetricConfig.GetMetricName = nil
	return cfg
}

func newFactory(componentType component.Type) exporter.Factory {
	return exporter.NewFactory(
		componentType,
		func() component.Config { return defaultConfig() },
	)
}

// testExporterConfig implements exporter.Config so we can test parsing of configuration.
type testExporterConfig struct {
	collector.Config `mapstructure:",squash"`
}

func defaultConfig() *testExporterConfig {
	return &testExporterConfig{
		Config: collector.DefaultConfig(),
	}
}
