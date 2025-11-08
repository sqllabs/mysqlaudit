// Copyright (C) 2025 JustCoding247. All rights reserved.

package server

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/sqllabs/sqlaudit/mysql"
	"github.com/sqllabs/sqlaudit/util/arena"
)

func newTestClientConn(buffer *bytes.Buffer) *clientConn {
	return &clientConn{
		alloc: arena.NewAllocator(32 * 1024),
		pkt: &packetIO{
			bufWriter: bufio.NewWriter(buffer),
		},
	}
}

func TestWriteAuthSwitchRequestNative(t *testing.T) {
	buf := new(bytes.Buffer)
	cc := newTestClientConn(buf)
	cc.salt = []byte("01234567890123456789")

	if err := cc.writeAuthSwitchRequest(mysql.AuthNativePassword); err != nil {
		t.Fatalf("writeAuthSwitchRequest failed: %v", err)
	}
	payload := buf.Bytes()
	if len(payload) <= 4 {
		t.Fatalf("unexpected packet length: %d", len(payload))
	}
	payload = payload[4:]
	if payload[0] != mysql.AuthSwitchRequestHeader {
		t.Fatalf("unexpected header: %x", payload[0])
	}
	plugin := payload[1 : 1+len(mysql.AuthNativePassword)]
	if string(plugin) != mysql.AuthNativePassword {
		t.Fatalf("unexpected plugin: %s", plugin)
	}
	pos := 1 + len(mysql.AuthNativePassword)
	if payload[pos] != 0 {
		t.Fatalf("expected plugin terminator")
	}
	pos++
	if !bytes.Equal(cc.salt, payload[pos:pos+len(cc.salt)]) {
		t.Fatalf("salt mismatch")
	}
	pos += len(cc.salt)
	if payload[pos] != 0 {
		t.Fatalf("native plugin should end with terminator")
	}
}

func TestWriteAuthSwitchRequestCaching(t *testing.T) {
	buf := new(bytes.Buffer)
	cc := newTestClientConn(buf)
	cc.salt = []byte("01234567890123456789")

	if err := cc.writeAuthSwitchRequest(mysql.AuthCachingSha2Password); err != nil {
		t.Fatalf("writeAuthSwitchRequest failed: %v", err)
	}
	payload := buf.Bytes()
	if len(payload) <= 4 {
		t.Fatalf("unexpected packet length: %d", len(payload))
	}
	payload = payload[4:]
	if payload[0] != mysql.AuthSwitchRequestHeader {
		t.Fatalf("unexpected header: %x", payload[0])
	}
	plugin := payload[1 : 1+len(mysql.AuthCachingSha2Password)]
	if string(plugin) != mysql.AuthCachingSha2Password {
		t.Fatalf("unexpected plugin: %s", plugin)
	}
	pos := 1 + len(mysql.AuthCachingSha2Password)
	if payload[pos] != 0 {
		t.Fatalf("expected plugin terminator")
	}
	pos++
	if !bytes.Equal(cc.salt, payload[pos:pos+len(cc.salt)]) {
		t.Fatalf("salt mismatch")
	}
	pos += len(cc.salt)
	if pos != len(payload) {
		t.Fatalf("caching sha2 should not append terminator")
	}
}
