package socks5_test

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/proxy"

	"github.com/JC5LZiy3HVfV5ux/socks5"
)

func TestProxyConnect(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		RemoteServerAddr  string
		ProxyAddr         string
		ProxyOpts         *socks5.Options
		DestinationUrl    string
		ClientCredentials *proxy.Auth
		UseTLS            bool
		Wait              []byte
		Error             string
	}{
		"IPv4_address": {
			RemoteServerAddr: "127.0.0.1:9451",
			ProxyAddr:        "127.0.0.1:1151",
			ProxyOpts: &socks5.Options{
				ListenAddr: "127.0.0.1:1151",
			},
			DestinationUrl: "http://127.0.0.1:9451/ping",
			Wait:           []byte("pong!"),
		},
		"FQDN_address": {
			RemoteServerAddr: "localhost:9452",
			ProxyAddr:        "127.0.0.1:1152",
			ProxyOpts: &socks5.Options{
				ListenAddr: "127.0.0.1:1152",
			},
			DestinationUrl: "http://localhost:9452/ping",
			Wait:           []byte("pong!"),
		},
		"authenticate_by_username_password": {
			RemoteServerAddr: "127.0.0.1:9453",
			ProxyAddr:        "127.0.0.1:1153",
			ProxyOpts: &socks5.Options{
				UserPassAuth: true,
				ListenAddr:   "127.0.0.1:1153",
			},
			DestinationUrl: "http://localhost:9453/ping",
			ClientCredentials: &proxy.Auth{
				User:     "root",
				Password: "password",
			},
			Wait: []byte("pong!"),
		},
		"wrong_authenticate_by_username_password": {
			ProxyAddr: "127.0.0.1:1154",
			ProxyOpts: &socks5.Options{
				StaticCredentials: map[string]string{
					"root": "password123",
				},
				UserPassAuth: true,
				ListenAddr:   "127.0.0.1:1154",
			},
			DestinationUrl: "http://localhost:9454",
			ClientCredentials: &proxy.Auth{
				User:     "root",
				Password: "password",
			},
			Error: "Get \"http://localhost:9454\": socks connect " +
				"tcp 127.0.0.1:1154->localhost:9454: username/password authentication failed",
		},
		"over_tls": {
			RemoteServerAddr: "127.0.0.1:9455",
			ProxyAddr:        "127.0.0.1:1155",
			ProxyOpts: &socks5.Options{
				UserPassAuth: true,
				ListenAddr:   "127.0.0.1:1155",
			},
			UseTLS:         true,
			DestinationUrl: "https://localhost:9455/ping",
			ClientCredentials: &proxy.Auth{
				User:     "root",
				Password: "password",
			},
			Wait: []byte("pong!"),
		},
		"connection_refused": {
			ProxyAddr: "127.0.0.1:1156",
			ProxyOpts: &socks5.Options{
				Logger:     &socks5.NoOutputLogger{},
				ListenAddr: "127.0.0.1:1156",
			},
			DestinationUrl: "http://localhost:4000",
			Error: "Get \"http://localhost:4000\": socks connect " +
				"tcp 127.0.0.1:1156->localhost:4000: unknown error connection refused",
		},
		"IPv6_address": {
			RemoteServerAddr: "[::1]:9457",
			ProxyAddr:        "127.0.0.1:1157",
			ProxyOpts: &socks5.Options{
				ListenAddr: "127.0.0.1:1157",
			},
			DestinationUrl: "http://[::1]:9457/ping",
			Wait:           []byte("pong!"),
		},
		"host_unreachable": {
			ProxyAddr: "127.0.0.1:1158",
			ProxyOpts: &socks5.Options{
				Logger:     &socks5.NoOutputLogger{},
				ListenAddr: "127.0.0.1:1158",
			},
			DestinationUrl: "http://no_such_host.test",
			Error: "Get \"http://no_such_host.test\": socks connect " +
				"tcp 127.0.0.1:1158->no_such_host.test:80: unknown error host unreachable",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			go runRemoteServer(tc.RemoteServerAddr, tc.UseTLS)
			go runProxy(tc.ProxyOpts, tc.UseTLS)

			// Wait for socks5 proxy to start
			time.Sleep(100 * time.Millisecond)

			client, err := setupClient(tc.ProxyAddr, tc.ClientCredentials)
			assert.NoError(t, err)

			resp, err := client.Get(tc.DestinationUrl)
			if err != nil {
				require.EqualError(t, err, tc.Error)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			require.Equal(t, tc.Wait, body)
		})
	}
}
