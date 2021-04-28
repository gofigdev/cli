package main

import (
	"context"
	"os"
	"testing"
)

var ctx = context.Background()

func TestGOPROXYValues(t *testing.T) {
	for _, tc := range []struct {
		name     string
		goproxy  string
		proxyURL string
		want     string
	}{
		{
			name:     "happy path",
			goproxy:  "proxy.golang.org,direct",
			proxyURL: "https://example.mygoproxy.com",
			want:     "GOPROXY=https://example.mygoproxy.com,proxy.golang.org,direct",
		},
		{
			name:     "pipe",
			goproxy:  "proxy.golang.org|direct",
			proxyURL: "https://example.mygoproxy.com",
			want:     "GOPROXY=https://example.mygoproxy.com,proxy.golang.org|direct",
		},
		{
			name:     "no goproxy value",
			goproxy:  "",
			proxyURL: "https://example.mygoproxy.com",
			want:     "GOPROXY=https://example.mygoproxy.com,https://proxy.golang.org,direct",
		},
		{
			name:     "already exists",
			goproxy:  "proxy.golang.org|https://example.mygoproxy.com",
			proxyURL: "https://example.mygoproxy.com",
			want:     "GOPROXY=proxy.golang.org|https://example.mygoproxy.com",
		},
		{
			name:     "empty goproxy",
			goproxy:  " ",
			proxyURL: "https://myproxy.com",
			want:     "GOPROXY=https://myproxy.com",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, "GOPROXY", tc.goproxy)
			got := getGOPROXYValue(ctx, tc.proxyURL)
			if got != tc.want {
				t.Fatalf("expected %q but got %q", tc.want, got)
			}
		})
	}
}

func TestGONOSUMDBValues(t *testing.T) {
	for _, tc := range []struct {
		name         string
		gonosumdb    string
		privatePaths []string
		want         string
	}{
		{
			name:         "happy path",
			gonosumdb:    "other.stuff/*",
			privatePaths: []string{"cli.gofig.dev/*"},
			want:         "GONOSUMDB=other.stuff/*,cli.gofig.dev/*",
		},
		{
			name:         "already exists",
			gonosumdb:    "other.stuff/*,cli.gofig.dev/*,more.stuff/*",
			privatePaths: []string{"private.stuff/*", "cli.gofig.dev/*"},
			want:         "GONOSUMDB=other.stuff/*,cli.gofig.dev/*,more.stuff/*,private.stuff/*",
		},
		{
			name:         "empty gonosumdb",
			gonosumdb:    " ",
			privatePaths: []string{"github.com/gofigdev/*"},
			want:         "GONOSUMDB=github.com/gofigdev/*",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, "GONOSUMDB", tc.gonosumdb)
			got := getGONOSUMDBValue(ctx, tc.privatePaths)
			if got != tc.want {
				t.Fatalf("expected %q but got %q", tc.want, got)
			}
		})
	}
}

func TestGOPRIVATEValues(t *testing.T) {
	for _, tc := range []struct {
		name         string
		goprivate    string
		privatePaths []string
		want         string
	}{
		{
			name:         "happy path",
			goprivate:    "other.stuff/*",
			privatePaths: []string{"cli.gofig.dev/*"},
			want:         "GOPRIVATE=other.stuff/*",
		},
		{
			name:         "already exists",
			goprivate:    "other.stuff/*,cli.gofig.dev/*",
			privatePaths: []string{"cli.gofig.dev/*"},
			want:         "GOPRIVATE=other.stuff/*",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, "GOPRIVATE", tc.goprivate)
			got := getGOPRIVATEValue(ctx, tc.privatePaths)
			if got != tc.want {
				t.Fatalf("expected %q but got %q", tc.want, got)
			}
		})
	}
}

func setEnv(t *testing.T, key, val string) {
	current := os.Getenv(key)
	t.Cleanup(func() {
		os.Setenv(key, current)
	})
	os.Setenv(key, val)
}
