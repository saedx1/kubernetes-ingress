package configs

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

func TestParseConfigMapWithAppProtectCompressedRequestsAction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		action string
		expect string
		msg    string
	}{
		{
			action: "pass",
			expect: "pass",
			msg:    "valid action pass",
		},
		{
			action: "drop",
			expect: "drop",
			msg:    "valid action drop",
		},
		{
			action: "invalid",
			expect: "",
			msg:    "invalid action",
		},
		{
			action: "",
			expect: "",
			msg:    "empty action",
		},
	}
	nginxPlus := true
	hasAppProtect := true
	hasAppProtectDos := false
	hasTLSPassthrough := false
	for _, test := range tests {
		cm := &v1.ConfigMap{
			Data: map[string]string{
				"app-protect-compressed-requests-action": test.action,
			},
		}
		result, _ := ParseConfigMap(context.Background(), cm, nginxPlus, hasAppProtect, hasAppProtectDos, hasTLSPassthrough, makeEventLogger())
		if result.MainAppProtectCompressedRequestsAction != test.expect {
			t.Errorf("ParseConfigMap() returned %q but expected %q for the case %s", result.MainAppProtectCompressedRequestsAction, test.expect, test.msg)
		}
	}
}

func TestParseConfigMapWithAppProtectReconnectPeriod(t *testing.T) {
	tests := []struct {
		period string
		expect string
		msg    string
	}{
		{
			period: "25",
			expect: "25",
			msg:    "valid period 25",
		},
		{
			period: "13.875",
			expect: "13.875",
			msg:    "valid period 13.875",
		},
		{
			period: "0.125",
			expect: "0.125",
			msg:    "valid period 0.125",
		},
		{
			period: "60",
			expect: "60",
			msg:    "valid period 60",
		},
		{
			period: "60.1",
			expect: "",
			msg:    "invalid period 60.1",
		},
		{
			period: "100",
			expect: "",
			msg:    "invalid period 100",
		},
		{
			period: "0",
			expect: "",
			msg:    "invalid period 0",
		},
		{
			period: "-5",
			expect: "",
			msg:    "invalid period -5",
		},
		{
			period: "",
			expect: "",
			msg:    "empty period",
		},
	}
	nginxPlus := true
	hasAppProtect := true
	hasAppProtectDos := false
	hasTLSPassthrough := false
	for _, test := range tests {
		cm := &v1.ConfigMap{
			Data: map[string]string{
				"app-protect-reconnect-period-seconds": test.period,
			},
		}
		result, _ := ParseConfigMap(context.Background(), cm, nginxPlus, hasAppProtect, hasAppProtectDos, hasTLSPassthrough, makeEventLogger())
		if result.MainAppProtectReconnectPeriod != test.expect {
			t.Errorf("ParseConfigMap() returned %q but expected %q for the case %s", result.MainAppProtectReconnectPeriod, test.expect, test.msg)
		}
	}
}

func TestParseConfigMapWithTLSPassthroughProxyProtocol(t *testing.T) {
	t.Parallel()
	tests := []struct {
		realIPheader string
		want         string
		msg          string
	}{
		{
			realIPheader: "proxy_protocol",
			want:         "",
			msg:          "valid realIPheader proxy_protocol, ignored when TLS Passthrough is enabled",
		},
		{
			realIPheader: "X-Forwarded-For",
			want:         "",
			msg:          "invalid realIPheader X-Forwarded-For, ignored when TLS Passthrough is enabled",
		},
		{
			realIPheader: "",
			want:         "",
			msg:          "empty real-ip-header",
		},
	}
	nginxPlus := true
	hasAppProtect := true
	hasAppProtectDos := false
	hasTLSPassthrough := true
	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			cm := &v1.ConfigMap{
				Data: map[string]string{
					"real-ip-header": test.realIPheader,
				},
			}
			result, _ := ParseConfigMap(context.Background(), cm, nginxPlus, hasAppProtect, hasAppProtectDos, hasTLSPassthrough, makeEventLogger())
			if result.RealIPHeader != test.want {
				t.Errorf("want %q, got %q", test.want, result.RealIPHeader)
			}
		})
	}
}

func TestParseConfigMapWithoutTLSPassthroughProxyProtocol(t *testing.T) {
	t.Parallel()
	tests := []struct {
		realIPheader string
		want         string
		msg          string
	}{
		{
			realIPheader: "proxy_protocol",
			want:         "proxy_protocol",
			msg:          "valid real-ip-header proxy_protocol",
		},
		{
			realIPheader: "X-Forwarded-For",
			want:         "X-Forwarded-For",
			msg:          "valid real-ip-header X-Forwarded-For",
		},
		{
			realIPheader: "",
			want:         "",
			msg:          "empty real-ip-header",
		},
	}
	nginxPlus := true
	hasAppProtect := true
	hasAppProtectDos := false
	hasTLSPassthrough := false
	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			cm := &v1.ConfigMap{
				Data: map[string]string{
					"real-ip-header": test.realIPheader,
				},
			}
			result, _ := ParseConfigMap(context.Background(), cm, nginxPlus, hasAppProtect, hasAppProtectDos, hasTLSPassthrough, makeEventLogger())
			if result.RealIPHeader != test.want {
				t.Errorf("want %q, got %q", test.want, result.RealIPHeader)
			}
		})
	}
}

func TestParseConfigMapAccessLog(t *testing.T) {
	t.Parallel()
	tests := []struct {
		accessLog    string
		accessLogOff string
		want         string
		msg          string
	}{
		{
			accessLogOff: "False",
			accessLog:    "syslog:server=localhost:514",
			want:         "syslog:server=localhost:514",
			msg:          "Non default access_log",
		},
		{
			accessLogOff: "False",
			accessLog:    "/tmp/nginx main",
			want:         "/dev/stdout main",
			msg:          "access_log to file is not allowed",
		},
		{
			accessLogOff: "True",
			accessLog:    "/dev/stdout main",
			want:         "off",
			msg:          "Disabled access_log",
		},
	}
	nginxPlus := true
	hasAppProtect := false
	hasAppProtectDos := false
	hasTLSPassthrough := false
	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			cm := &v1.ConfigMap{
				Data: map[string]string{
					"access-log":     test.accessLog,
					"access-log-off": test.accessLogOff,
				},
			}
			result, _ := ParseConfigMap(context.Background(), cm, nginxPlus, hasAppProtect, hasAppProtectDos, hasTLSPassthrough, makeEventLogger())
			if result.MainAccessLog != test.want {
				t.Errorf("want %q, got %q", test.want, result.MainAccessLog)
			}
		})
	}
}

func TestParseConfigMapAccessLogDefault(t *testing.T) {
	t.Parallel()
	tests := []struct {
		accessLog    string
		accessLogOff string
		want         string
		msg          string
	}{
		{
			want: "/dev/stdout main",
			msg:  "Default access_log",
		},
	}
	nginxPlus := true
	hasAppProtect := false
	hasAppProtectDos := false
	hasTLSPassthrough := false
	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			cm := &v1.ConfigMap{
				Data: map[string]string{
					"access-log-off": "False",
				},
			}
			result, _ := ParseConfigMap(context.Background(), cm, nginxPlus, hasAppProtect, hasAppProtectDos, hasTLSPassthrough, makeEventLogger())
			if result.MainAccessLog != test.want {
				t.Errorf("want %q, got %q", test.want, result.MainAccessLog)
			}
		})
	}
}

func TestParseMGMTConfigMapError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		configMap *v1.ConfigMap
		msg       string
	}{
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "",
				},
			},
			msg: "Must have license-token-secret-name",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{},
			},
			msg: "Must have license-token-secret-name key",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			_, _, err := ParseMGMTConfigMap(context.Background(), test.configMap, makeEventLogger())

			if err == nil {
				t.Errorf("Expected error, got nil")
			}
		})
	}
}

func TestParseMGMTConfigMapWarnings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		configMap *v1.ConfigMap
		msg       string
	}{
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"enforce-initial-report":    "7",
				},
			},
			msg: "enforce-initial-report is invalid",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"enforce-initial-report":    "",
				},
			},
			msg: "enforce-initial-report set empty",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"usage-report-interval":     "",
				},
			},
			msg: "usage-report-interval set empty",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"usage-report-interval":     "1s",
				},
			},
			msg: "usage-report-interval set below allowed value",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"usage-report-interval":     "1s",
				},
			},
			msg: "usage-report-interval set below allowed value",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"ssl-verify":                "10",
				},
			},
			msg: "ssl-verify set to an invalid int",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"ssl-verify":                "test",
				},
			},
			msg: "ssl-verify set to an invalid value",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"ssl-verify":                "",
				},
			},
			msg: "ssl-verify set to an empty string",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"resolver-ipv6":             "",
				},
			},
			msg: "resolver-ipv6 set to an empty string",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"resolver-ipv6":             "10",
				},
			},
			msg: "resolver-ipv6 set to an invalid int",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			_, configWarnings, err := ParseMGMTConfigMap(context.Background(), test.configMap, makeEventLogger())
			if err != nil {
				t.Errorf("expected nil, got err: %v", err)
			}
			if !configWarnings {
				t.Fatal("Expected warnings, got none")
			}
		})
	}
}

func TestParseMGMTConfigMapLicense(t *testing.T) {
	t.Parallel()
	test := struct {
		configMap *v1.ConfigMap
		want      *MGMTConfigParams
		msg       string
	}{
		configMap: &v1.ConfigMap{
			Data: map[string]string{
				"license-token-secret-name": "license-token",
			},
		},
		want: &MGMTConfigParams{
			Secrets: MGMTSecrets{
				License: "license-token",
			},
		},
		msg: "Has only license-token-secret-name",
	}

	t.Run(test.msg, func(t *testing.T) {
		result, warnings, err := ParseMGMTConfigMap(context.Background(), test.configMap, makeEventLogger())
		if err != nil {
			t.Fatal(err)
		}
		if warnings {
			t.Fatal("Unexpected warnings")
		}
		if result.Secrets.License != test.want.Secrets.License {
			t.Errorf("LicenseTokenSecretNane: want %q, got %q", test.want.Secrets.License, result.Secrets.License)
		}
	})
}

func TestParseMGMTConfigMapEnforceInitialReport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		configMap *v1.ConfigMap
		want      *MGMTConfigParams
		msg       string
	}{
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"enforce-initial-report":    "false",
				},
			},
			want: &MGMTConfigParams{
				EnforceInitialReport: BoolToPointerBool(false),
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "enforce-initial-report set to false",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"enforce-initial-report":    "true",
				},
			},
			want: &MGMTConfigParams{
				EnforceInitialReport: BoolToPointerBool(true),
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "enforce-initial-report set to true",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			result, warnings, err := ParseMGMTConfigMap(context.Background(), test.configMap, makeEventLogger())
			if err != nil {
				t.Fatal(err)
			}
			if warnings {
				t.Error("Unexpected warnings")
			}

			if result.EnforceInitialReport == nil {
				t.Errorf("EnforceInitialReport: want %v, got nil", *test.want.EnforceInitialReport)
			}
			if *result.EnforceInitialReport != *test.want.EnforceInitialReport {
				t.Errorf("EnforceInitialReport: want %v, got %v", *test.want.EnforceInitialReport, *result.EnforceInitialReport)
			}
		})
	}
}

func TestParseMGMTConfigMapSSLVerify(t *testing.T) {
	t.Parallel()
	tests := []struct {
		configMap *v1.ConfigMap
		want      *MGMTConfigParams
		msg       string
	}{
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"ssl-verify":                "false",
				},
			},
			want: &MGMTConfigParams{
				SSLVerify: BoolToPointerBool(false),
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "ssl-verify set to false",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"ssl-verify":                "true",
				},
			},
			want: &MGMTConfigParams{
				SSLVerify: BoolToPointerBool(true),
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "ssl-verify set to true",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			result, warnings, err := ParseMGMTConfigMap(context.Background(), test.configMap, makeEventLogger())
			if err != nil {
				t.Fatal(err)
			}
			if warnings {
				t.Error("Unexpected warnings")
			}

			if result.SSLVerify == nil {
				t.Errorf("ssl-verify: want %v, got nil", *test.want.SSLVerify)
			}
			if *result.SSLVerify != *test.want.SSLVerify {
				t.Errorf("ssl-verify: want %v, got %v", *test.want.SSLVerify, *result.SSLVerify)
			}
		})
	}
}

func TestParseMGMTConfigMapUsageReportInterval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		configMap *v1.ConfigMap
		want      *MGMTConfigParams
		msg       string
	}{
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"usage-report-interval":     "120s",
				},
			},
			want: &MGMTConfigParams{
				Interval: "120s",
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "usage report interval set to 120s",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"usage-report-interval":     "20m",
				},
			},
			want: &MGMTConfigParams{
				Interval: "20m",
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "usage report interval set to 20m",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"usage-report-interval":     "1h",
				},
			},
			want: &MGMTConfigParams{
				Interval: "1h",
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "usage report interval set to 1h",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"usage-report-interval":     "24h",
				},
			},
			want: &MGMTConfigParams{
				Interval: "24h",
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "usage report interval set to 24h",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			result, warnings, err := ParseMGMTConfigMap(context.Background(), test.configMap, makeEventLogger())
			if err != nil {
				t.Fatal(err)
			}
			if warnings {
				t.Error("Unexpected warnings")
			}

			if result.Interval == "" {
				t.Errorf("UsageReportInterval: want %s, got empty string", test.want.Interval)
			}
			if result.Interval != test.want.Interval {
				t.Errorf("UsageReportInterval: want %v, got %v", test.want.Interval, result.Interval)
			}
		})
	}
}

func TestParseMGMTConfigMapResolverIPV6(t *testing.T) {
	t.Parallel()
	tests := []struct {
		configMap *v1.ConfigMap
		want      *MGMTConfigParams
		msg       string
	}{
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"resolver-ipv6":             "false",
				},
			},
			want: &MGMTConfigParams{
				ResolverIPV6: BoolToPointerBool(false),
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "resolver-ipv6 set to false",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"resolver-ipv6":             "true",
				},
			},
			want: &MGMTConfigParams{
				ResolverIPV6: BoolToPointerBool(true),
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "resolver-ipv6 set to true",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			result, warnings, err := ParseMGMTConfigMap(context.Background(), test.configMap, makeEventLogger())
			if err != nil {
				t.Fatal(err)
			}
			if warnings {
				t.Error("Unexpected warnings")
			}

			if result.ResolverIPV6 == nil {
				t.Errorf("resolver-ipv6: want %v, got nil", *test.want.ResolverIPV6)
			}
			if *result.ResolverIPV6 != *test.want.ResolverIPV6 {
				t.Errorf("resolver-ipv6: want %v, got %v", *test.want.ResolverIPV6, *result.ResolverIPV6)
			}
		})
	}
}

func TestParseMGMTConfigMapUsageReportEndpoint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		configMap *v1.ConfigMap
		want      *MGMTConfigParams
		msg       string
	}{
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"usage-report-endpoint":     "product.connect.nginx.com",
				},
			},
			want: &MGMTConfigParams{
				Endpoint: "product.connect.nginx.com",
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "usage report endpoint set to product.connect.nginx.com",
		},
		{
			configMap: &v1.ConfigMap{
				Data: map[string]string{
					"license-token-secret-name": "license-token",
					"usage-report-endpoint":     "product.connect.nginx.com:80",
				},
			},
			want: &MGMTConfigParams{
				Endpoint: "product.connect.nginx.com:80",
				Secrets: MGMTSecrets{
					License: "license-token",
				},
			},
			msg: "usage report endpoint set to product.connect.nginx.com with port 80",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			result, warnings, err := ParseMGMTConfigMap(context.Background(), test.configMap, makeEventLogger())
			if err != nil {
				t.Fatal(err)
			}
			if warnings {
				t.Error("Unexpected warnings")
			}

			if result.Endpoint == "" {
				t.Errorf("UsageReportEndpoint: want %s, got empty string", test.want.Endpoint)
			}
			if result.Endpoint != test.want.Endpoint {
				t.Errorf("UsageReportEndpoint: want %v, got %v", test.want.Endpoint, result.Endpoint)
			}
		})
	}
}

func makeEventLogger() record.EventRecorder {
	return record.NewFakeRecorder(1024)
}
