// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/api/v1alpha1"
)

var (
	TLSSecretKind       = v1.Kind("Secret")
	TLSUnrecognizedKind = v1.Kind("Unrecognized")
)

func TestValidateEnvoyGateway(t *testing.T) {
	eg := v1alpha1.DefaultEnvoyGateway()

	testCases := []struct {
		name   string
		eg     *v1alpha1.EnvoyGateway
		expect bool
	}{
		{
			name:   "default",
			eg:     eg,
			expect: true,
		},
		{
			name:   "unspecified envoy gateway",
			eg:     nil,
			expect: false,
		},
		{
			name: "unspecified gateway",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: false,
		},
		{
			name: "unspecified provider",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway: v1alpha1.DefaultGateway(),
				},
			},
			expect: false,
		},
		{
			name: "empty gateway controllerName",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  &v1alpha1.Gateway{ControllerName: ""},
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: false,
		},
		{
			name: "unsupported provider",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: &v1alpha1.EnvoyGatewayProvider{Type: v1alpha1.ProviderTypeFile},
				},
			},
			expect: false,
		},
		{
			name: "empty ratelimit",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:   v1alpha1.DefaultGateway(),
					Provider:  v1alpha1.DefaultEnvoyGatewayProvider(),
					RateLimit: &v1alpha1.RateLimit{},
				},
			},
			expect: false,
		},
		{
			name: "empty ratelimit redis setting",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					RateLimit: &v1alpha1.RateLimit{
						Backend: v1alpha1.RateLimitDatabaseBackend{
							Type:  v1alpha1.RedisBackendType,
							Redis: &v1alpha1.RateLimitRedisSettings{},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "unknown ratelimit redis url format",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					RateLimit: &v1alpha1.RateLimit{
						Backend: v1alpha1.RateLimitDatabaseBackend{
							Type: v1alpha1.RedisBackendType,
							Redis: &v1alpha1.RateLimitRedisSettings{
								URL: ":foo",
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "happy ratelimit redis settings",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					RateLimit: &v1alpha1.RateLimit{
						Backend: v1alpha1.RateLimitDatabaseBackend{
							Type: v1alpha1.RedisBackendType,
							Redis: &v1alpha1.RateLimitRedisSettings{
								URL: "localhost:6376",
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "happy extension settings",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Resources: []v1alpha1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &v1alpha1.ExtensionHooks{
							XDSTranslator: &v1alpha1.XDSTranslatorHooks{
								Pre: []v1alpha1.XDSTranslatorHook{},
								Post: []v1alpha1.XDSTranslatorHook{
									v1alpha1.XDSHTTPListener,
									v1alpha1.XDSTranslation,
									v1alpha1.XDSRoute,
									v1alpha1.XDSVirtualHost,
								},
							},
						},
						Service: &v1alpha1.ExtensionService{
							Host: "foo.extension",
							Port: 80,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "happy extension settings tls",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Resources: []v1alpha1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &v1alpha1.ExtensionHooks{
							XDSTranslator: &v1alpha1.XDSTranslatorHooks{
								Pre: []v1alpha1.XDSTranslatorHook{},
								Post: []v1alpha1.XDSTranslatorHook{
									v1alpha1.XDSHTTPListener,
									v1alpha1.XDSTranslation,
									v1alpha1.XDSRoute,
									v1alpha1.XDSVirtualHost,
								},
							},
						},
						Service: &v1alpha1.ExtensionService{
							Host: "foo.extension",
							Port: 443,
							TLS: &v1alpha1.ExtensionTLS{
								CertificateRef: v1.SecretObjectReference{
									Kind: &TLSSecretKind,
									Name: v1.ObjectName("certificate"),
								},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "happy extension settings no resources",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Hooks: &v1alpha1.ExtensionHooks{
							XDSTranslator: &v1alpha1.XDSTranslatorHooks{
								Pre: []v1alpha1.XDSTranslatorHook{},
								Post: []v1alpha1.XDSTranslatorHook{
									v1alpha1.XDSHTTPListener,
									v1alpha1.XDSTranslation,
									v1alpha1.XDSRoute,
									v1alpha1.XDSVirtualHost,
								},
							},
						},
						Service: &v1alpha1.ExtensionService{
							Host: "foo.extension",
							Port: 443,
							TLS: &v1alpha1.ExtensionTLS{
								CertificateRef: v1.SecretObjectReference{
									Kind: &TLSSecretKind,
									Name: v1.ObjectName("certificate"),
								},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "unknown TLS certificateRef in extension settings",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Resources: []v1alpha1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &v1alpha1.ExtensionHooks{
							XDSTranslator: &v1alpha1.XDSTranslatorHooks{
								Pre: []v1alpha1.XDSTranslatorHook{},
								Post: []v1alpha1.XDSTranslatorHook{
									v1alpha1.XDSHTTPListener,
									v1alpha1.XDSTranslation,
									v1alpha1.XDSRoute,
									v1alpha1.XDSVirtualHost,
								},
							},
						},
						Service: &v1alpha1.ExtensionService{
							Host: "foo.extension",
							Port: 8080,
							TLS: &v1alpha1.ExtensionTLS{
								CertificateRef: v1.SecretObjectReference{
									Kind: &TLSUnrecognizedKind,
									Name: v1.ObjectName("certificate"),
								},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "empty service in extension settings",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Resources: []v1alpha1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &v1alpha1.ExtensionHooks{
							XDSTranslator: &v1alpha1.XDSTranslatorHooks{
								Pre: []v1alpha1.XDSTranslatorHook{},
								Post: []v1alpha1.XDSTranslatorHook{
									v1alpha1.XDSHTTPListener,
									v1alpha1.XDSTranslation,
									v1alpha1.XDSRoute,
									v1alpha1.XDSVirtualHost,
								},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "empty hooks in extension settings",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Resources: []v1alpha1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Service: &v1alpha1.ExtensionService{
							Host: "foo.extension",
							Port: 8080,
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "valid gateway logging level info",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Logging: &v1alpha1.EnvoyGatewayLogging{
						Level: map[v1alpha1.EnvoyGatewayLogComponent]v1alpha1.LogLevel{
							v1alpha1.LogComponentGatewayDefault: v1alpha1.LogLevelInfo,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "valid gateway logging level warn",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Logging: &v1alpha1.EnvoyGatewayLogging{
						Level: map[v1alpha1.EnvoyGatewayLogComponent]v1alpha1.LogLevel{
							v1alpha1.LogComponentGatewayDefault: v1alpha1.LogLevelWarn,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "valid gateway logging level error",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Logging: &v1alpha1.EnvoyGatewayLogging{
						Level: map[v1alpha1.EnvoyGatewayLogComponent]v1alpha1.LogLevel{
							v1alpha1.LogComponentGatewayDefault: v1alpha1.LogLevelError,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "valid gateway logging level debug",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Logging: &v1alpha1.EnvoyGatewayLogging{
						Level: map[v1alpha1.EnvoyGatewayLogComponent]v1alpha1.LogLevel{
							v1alpha1.LogComponentGatewayDefault: v1alpha1.LogLevelDebug,
							v1alpha1.LogComponentProviderRunner: v1alpha1.LogLevelDebug,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "invalid gateway logging level",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Logging: &v1alpha1.EnvoyGatewayLogging{
						Level: map[v1alpha1.EnvoyGatewayLogComponent]v1alpha1.LogLevel{
							v1alpha1.LogComponentGatewayDefault: "inffo",
						},
					},
				},
			},
			expect: false,
		}, {
			name: "valid gateway metrics sink",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Telemetry: &v1alpha1.EnvoyGatewayTelemetry{
						Metrics: &v1alpha1.EnvoyGatewayMetrics{
							Sinks: []v1alpha1.EnvoyGatewayMetricSink{
								{
									Type: v1alpha1.MetricSinkTypeOpenTelemetry,
									OpenTelemetry: &v1alpha1.EnvoyGatewayOpenTelemetrySink{
										Host:     "x.x.x.x",
										Port:     4317,
										Protocol: "grpc",
									},
								},
							},
						},
					},
				},
			},
			expect: true,
		}, {
			name: "invalid gateway metrics sink",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Telemetry: &v1alpha1.EnvoyGatewayTelemetry{
						Metrics: &v1alpha1.EnvoyGatewayMetrics{
							Sinks: []v1alpha1.EnvoyGatewayMetricSink{
								{
									Type: v1alpha1.MetricSinkTypeOpenTelemetry,
								},
							},
						},
					},
				},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEnvoyGateway(tc.eg)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEnvoyGateway(t *testing.T) {
	envoyGateway := v1alpha1.DefaultEnvoyGateway()
	assert.True(t, envoyGateway.Provider != nil)
	assert.True(t, envoyGateway.Gateway != nil)
	assert.True(t, envoyGateway.Logging != nil)
	envoyGateway.SetEnvoyGatewayDefaults()
	assert.Equal(t, envoyGateway.Logging, v1alpha1.DefaultEnvoyGatewayLogging())

	logging := v1alpha1.DefaultEnvoyGatewayLogging()
	assert.True(t, logging != nil)
	assert.True(t, logging.Level[v1alpha1.LogComponentGatewayDefault] == v1alpha1.LogLevelInfo)

	gatewayLogging := &v1alpha1.EnvoyGatewayLogging{
		Level: logging.Level,
	}
	gatewayLogging.SetEnvoyGatewayLoggingDefaults()
	assert.True(t, gatewayLogging != nil)
	assert.True(t, gatewayLogging.Level[v1alpha1.LogComponentGatewayDefault] == v1alpha1.LogLevelInfo)
}

func TestDefaultEnvoyGatewayLoggingLevel(t *testing.T) {
	type args struct {
		component string
		level     v1alpha1.LogLevel
	}
	tests := []struct {
		name string
		args args
		want v1alpha1.LogLevel
	}{
		{
			name: "test default info level for empty level",
			args: args{component: "", level: ""},
			want: v1alpha1.LogLevelInfo,
		},
		{
			name: "test default info level for empty level",
			args: args{component: string(v1alpha1.LogComponentGatewayDefault), level: ""},
			want: v1alpha1.LogLevelInfo,
		},
		{
			name: "test default info level for info level",
			args: args{component: string(v1alpha1.LogComponentGatewayDefault), level: v1alpha1.LogLevelInfo},
			want: v1alpha1.LogLevelInfo,
		},
		{
			name: "test default error level for error level",
			args: args{component: string(v1alpha1.LogComponentGatewayDefault), level: v1alpha1.LogLevelError},
			want: v1alpha1.LogLevelError,
		},
		{
			name: "test gateway-api error level for error level",
			args: args{component: string(v1alpha1.LogComponentGatewayAPIRunner), level: v1alpha1.LogLevelError},
			want: v1alpha1.LogLevelError,
		},
		{
			name: "test gateway-api info level for info level",
			args: args{component: string(v1alpha1.LogComponentGatewayAPIRunner), level: v1alpha1.LogLevelInfo},
			want: v1alpha1.LogLevelInfo,
		},
		{
			name: "test default gateway-api warn level for info level",
			args: args{component: string(v1alpha1.LogComponentGatewayAPIRunner), level: ""},
			want: v1alpha1.LogLevelInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logging := &v1alpha1.EnvoyGatewayLogging{}
			if got := logging.DefaultEnvoyGatewayLoggingLevel(tt.args.level); got != tt.want {
				t.Errorf("defaultLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvoyGatewayProvider(t *testing.T) {
	envoyGateway := &v1alpha1.EnvoyGateway{
		TypeMeta:         metav1.TypeMeta{},
		EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{Provider: v1alpha1.DefaultEnvoyGatewayProvider()},
	}
	assert.True(t, envoyGateway.Provider != nil)

	envoyGatewayProvider := envoyGateway.GetEnvoyGatewayProvider()
	assert.True(t, envoyGatewayProvider.Kubernetes == nil)
	assert.Equal(t, envoyGateway.Provider, envoyGatewayProvider)

	envoyGatewayProvider.Kubernetes = v1alpha1.DefaultEnvoyGatewayKubeProvider()
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, v1alpha1.DefaultKubernetesDeployment(v1alpha1.DefaultRateLimitImage))

	envoyGatewayProvider.Kubernetes = &v1alpha1.EnvoyGatewayKubernetesProvider{}
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment == nil)

	envoyGatewayProvider.Kubernetes = &v1alpha1.EnvoyGatewayKubernetesProvider{
		RateLimitDeployment: &v1alpha1.KubernetesDeploymentSpec{
			Replicas:  nil,
			Pod:       nil,
			Container: nil,
		}}
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas == nil)
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod == nil)
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container == nil)
	envoyGatewayKubeProvider := envoyGatewayProvider.GetEnvoyGatewayKubeProvider()

	envoyGatewayProvider.Kubernetes = &v1alpha1.EnvoyGatewayKubernetesProvider{
		RateLimitDeployment: &v1alpha1.KubernetesDeploymentSpec{
			Replicas: nil,
			Pod:      nil,
			Container: &v1alpha1.KubernetesContainerSpec{
				Resources:       nil,
				SecurityContext: nil,
				Image:           nil,
			},
		}}
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources == nil)
	envoyGatewayProvider.GetEnvoyGatewayKubeProvider()

	assert.True(t, envoyGatewayProvider.Kubernetes != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes, envoyGatewayKubeProvider)

	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, v1alpha1.DefaultKubernetesDeployment(v1alpha1.DefaultRateLimitImage))
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas, v1alpha1.DefaultKubernetesDeploymentReplicas())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod, v1alpha1.DefaultKubernetesPod())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container, v1alpha1.DefaultKubernetesContainer(v1alpha1.DefaultRateLimitImage))
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources, v1alpha1.DefaultResourceRequirements())
	assert.True(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image != nil)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image, v1alpha1.DefaultKubernetesContainerImage(v1alpha1.DefaultRateLimitImage))
}

func TestEnvoyGatewayAdmin(t *testing.T) {
	// default envoygateway config admin should not be nil
	eg := v1alpha1.DefaultEnvoyGateway()
	assert.True(t, eg.Admin != nil)

	// get default admin config from envoygateway
	// values should be set in default
	egAdmin := eg.GetEnvoyGatewayAdmin()
	assert.True(t, egAdmin != nil)
	assert.True(t, egAdmin.Address.Port == v1alpha1.GatewayAdminPort)
	assert.True(t, egAdmin.Address.Host == v1alpha1.GatewayAdminHost)
	assert.True(t, egAdmin.EnableDumpConfig == false)
	assert.True(t, egAdmin.EnablePprof == false)

	// override the admin config
	// values should be updated
	eg.Admin = &v1alpha1.EnvoyGatewayAdmin{
		Address: &v1alpha1.EnvoyGatewayAdminAddress{
			Host: "0.0.0.0",
			Port: 19010,
		},
		EnableDumpConfig: true,
		EnablePprof:      true,
	}

	assert.True(t, eg.GetEnvoyGatewayAdmin().Address.Port == 19010)
	assert.True(t, eg.GetEnvoyGatewayAdmin().Address.Host == "0.0.0.0")
	assert.True(t, eg.GetEnvoyGatewayAdmin().EnableDumpConfig == true)
	assert.True(t, eg.GetEnvoyGatewayAdmin().EnablePprof == true)

	// set eg defaults when admin is nil
	// the admin should not be nil
	eg.Admin = nil
	eg.SetEnvoyGatewayDefaults()
	assert.True(t, eg.Admin != nil)
	assert.True(t, eg.Admin.Address.Port == v1alpha1.GatewayAdminPort)
	assert.True(t, eg.Admin.Address.Host == v1alpha1.GatewayAdminHost)
	assert.True(t, eg.Admin.EnableDumpConfig == false)
	assert.True(t, eg.Admin.EnablePprof == false)
}

func TestEnvoyGatewayTelemetry(t *testing.T) {
	// default envoygateway config telemetry should not be nil
	eg := v1alpha1.DefaultEnvoyGateway()
	assert.True(t, eg.Telemetry != nil)

	// get default telemetry config from envoygateway
	// values should be set in default
	egTelemetry := eg.GetEnvoyGatewayTelemetry()
	assert.True(t, egTelemetry != nil)
	assert.True(t, egTelemetry.Metrics != nil)
	assert.True(t, egTelemetry.Metrics.Prometheus.Disable == false)
	assert.True(t, egTelemetry.Metrics.Sinks == nil)

	// override the telemetry config
	// values should be updated
	eg.Telemetry.Metrics = &v1alpha1.EnvoyGatewayMetrics{
		Prometheus: &v1alpha1.EnvoyGatewayPrometheusProvider{
			Disable: true,
		},
		Sinks: []v1alpha1.EnvoyGatewayMetricSink{
			{
				Type: v1alpha1.MetricSinkTypeOpenTelemetry,
				OpenTelemetry: &v1alpha1.EnvoyGatewayOpenTelemetrySink{
					Host:     "otel-collector.monitoring.svc.cluster.local",
					Protocol: "grpc",
					Port:     4317,
				},
			}, {
				Type: v1alpha1.MetricSinkTypeOpenTelemetry,
				OpenTelemetry: &v1alpha1.EnvoyGatewayOpenTelemetrySink{
					Host:     "otel-collector.monitoring.svc.cluster.local",
					Protocol: "http",
					Port:     4318,
				},
			},
		},
	}

	assert.True(t, eg.GetEnvoyGatewayTelemetry().Metrics.Prometheus.Disable == true)
	assert.True(t, len(eg.GetEnvoyGatewayTelemetry().Metrics.Sinks) == 2)
	assert.True(t, eg.GetEnvoyGatewayTelemetry().Metrics.Sinks[0].Type == v1alpha1.MetricSinkTypeOpenTelemetry)

	// set eg defaults when telemetry is nil
	// the telemetry should not be nil
	eg.Telemetry = nil
	eg.SetEnvoyGatewayDefaults()
	assert.True(t, eg.Telemetry != nil)
	assert.True(t, eg.Telemetry.Metrics != nil)
	assert.True(t, eg.Telemetry.Metrics.Prometheus.Disable == false)
	assert.True(t, eg.Telemetry.Metrics.Sinks == nil)
}
