package socks5

import (
	"log"
	"os"
	"time"
)

type Options struct {
	ListenAddr          string            // default: 127.0.0.1:1080
	ReadTimeout         time.Duration     // default: none
	WriteTimeout        time.Duration     // default: none
	DialTimeout         time.Duration     // default: none
	ConnDatabaseTimeout time.Duration     // default: 15 seconds
	UserPassAuth        bool              // default: no authentication required
	StaticCredentials   map[string]string // default: root / password
	Logger              Logger            // default: stdoutLogger
	Store               Store             // default: defaultStore
	Driver              Driver            // default: defaultDriver
}

func (o Options) authMethods() map[byte]struct{} {
	methods := make(map[byte]struct{})

	switch {
	case o.UserPassAuth:
		methods[methodUsernamePassword] = struct{}{}
	default:
		methods[noAuthenticationRequired] = struct{}{}
	}

	return methods
}

func optsWithDefaults(opts *Options) *Options {
	if opts == nil {
		opts = &Options{}
	}

	if opts.ConnDatabaseTimeout == 0 {
		opts.ConnDatabaseTimeout = 15 * time.Second
	}

	if opts.Logger == nil {
		opts.Logger = &stdoutLogger{
			errorLogger: log.New(os.Stdout, "time: ", log.Ldate|log.Ltime),
		}
	}

	if opts.Store == nil {
		if opts.StaticCredentials == nil {
			opts.StaticCredentials = map[string]string{
				"root": "password",
			}
		}

		opts.Store = &defaultStore{
			db: opts.StaticCredentials,
		}
	}

	if opts.Driver == nil {
		if opts.ListenAddr == "" {
			opts.ListenAddr = "127.0.0.1:1080"
		}

		opts.Driver = &defaultDriver{
			listenAddr:  opts.ListenAddr,
			dialTimeout: opts.DialTimeout,
		}
	}

	return opts
}
