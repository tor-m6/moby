// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !go1.18
// +build !go1.18

package http2

import (
	"crypto/tls"
	"context"
	"net"
)

func tlsUnderlyingConn(tc *tls.Conn) net.Conn {
	return nil
}

// dialTLSWithContext opens a TLS connection.
func (t *Transport) dialTLSWithContext(ctx context.Context, network, addr string, cfg *tls.Config) (*tls.Conn, error) {
	cn, err := tls.Dial(network, addr, cfg)
	if err != nil {
		return nil, err
	}
	if err := cn.Handshake(); err != nil {
		return nil, err
	}
	if cfg.InsecureSkipVerify {
		return cn, nil
	}
	if err := cn.VerifyHostname(cfg.ServerName); err != nil {
		return nil, err
	}
	return cn, nil
}
