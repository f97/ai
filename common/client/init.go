package client

import (
	"fmt"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/env"
	"github.com/songquanpeng/one-api/common/logger"
	"net"
	"net/http"
	"net/url"
	"time"
)

var HTTPClient *http.Client
var ImpatientHTTPClient *http.Client
var UserContentRequestHTTPClient *http.Client

// getOptimizedTransport returns an optimized HTTP transport for single-user workload
func getOptimizedTransport(proxy *url.URL) *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyURL(proxy),
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(env.Int("HTTP_DIAL_TIMEOUT", 10)) * time.Second,
			KeepAlive: time.Duration(env.Int("HTTP_KEEPALIVE", 90)) * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   time.Duration(env.Int("HTTP_TLS_TIMEOUT", 10)) * time.Second,
		ResponseHeaderTimeout: time.Duration(env.Int("HTTP_RESPONSE_HEADER_TIMEOUT", 30)) * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// Optimized connection pooling for single-user
		MaxIdleConns:        env.Int("HTTP_MAX_IDLE_CONNS", 100),
		MaxIdleConnsPerHost: env.Int("HTTP_MAX_IDLE_CONNS_PER_HOST", 20),
		MaxConnsPerHost:     env.Int("HTTP_MAX_CONNS_PER_HOST", 50),
		IdleConnTimeout:     time.Duration(env.Int("HTTP_IDLE_CONN_TIMEOUT", 90)) * time.Second,
		ForceAttemptHTTP2:   true,
	}
}

func Init() {
	logger.SysLog("initializing HTTP clients with optimized settings")
	
	if config.UserContentRequestProxy != "" {
		logger.SysLog(fmt.Sprintf("using %s as proxy to fetch user content", config.UserContentRequestProxy))
		proxyURL, err := url.Parse(config.UserContentRequestProxy)
		if err != nil {
			logger.FatalLog(fmt.Sprintf("USER_CONTENT_REQUEST_PROXY set but invalid: %s", config.UserContentRequestProxy))
		}
		transport := getOptimizedTransport(proxyURL)
		UserContentRequestHTTPClient = &http.Client{
			Transport: transport,
			Timeout:   time.Second * time.Duration(config.UserContentRequestTimeout),
		}
	} else {
		transport := getOptimizedTransport(nil)
		UserContentRequestHTTPClient = &http.Client{
			Transport: transport,
			Timeout:   time.Second * time.Duration(config.UserContentRequestTimeout),
		}
	}
	
	var proxyURL *url.URL
	var transport *http.Transport
	if config.RelayProxy != "" {
		logger.SysLog(fmt.Sprintf("using %s as api relay proxy", config.RelayProxy))
		var err error
		proxyURL, err = url.Parse(config.RelayProxy)
		if err != nil {
			logger.FatalLog(fmt.Sprintf("RELAY_PROXY set but invalid: %s", config.RelayProxy))
		}
	}
	
	transport = getOptimizedTransport(proxyURL)

	if config.RelayTimeout == 0 {
		HTTPClient = &http.Client{
			Transport: transport,
		}
	} else {
		HTTPClient = &http.Client{
			Timeout:   time.Duration(config.RelayTimeout) * time.Second,
			Transport: transport,
		}
	}

	ImpatientHTTPClient = &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
	
	logger.SysLog(fmt.Sprintf("HTTP client optimizations: MaxIdleConns=%d, MaxIdleConnsPerHost=%d, KeepAlive=%ds", 
		transport.MaxIdleConns, transport.MaxIdleConnsPerHost, env.Int("HTTP_KEEPALIVE", 90)))
}
