package device

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestAWGJunkPacketsBounds(t *testing.T) {
	d := &Device{}
	d.junk.count.Store(3)
	d.junk.min.Store(64)
	d.junk.max.Store(64)
	pkts := d.JunkPackets()
	if len(pkts) != 3 {
		t.Fatalf("expected 3 packets, got %d", len(pkts))
	}
	for i, p := range pkts {
		if len(p) != 64 {
			t.Fatalf("packet %d len=%d, want 64", i, len(p))
		}
	}

	d.junk.count.Store(2)
	d.junk.min.Store(120)
	d.junk.max.Store(40)
	pkts2 := d.JunkPackets()
	if len(pkts2) != 2 {
		t.Fatalf("expected 2 packets, got %d", len(pkts2))
	}
	for i, p := range pkts2 {
		if len(p) != 120 {
			t.Fatalf("packet %d len=%d, want 120", i, len(p))
		}
	}

	d.junk.count.Store(2)
	d.junk.min.Store(0)
	d.junk.max.Store(0)
	if zeroPkts := d.JunkPackets(); len(zeroPkts) != 0 {
		t.Fatalf("expected 0 zero-sized packets, got %d", len(zeroPkts))
	}
}

func TestAWGCounterObfTag(t *testing.T) {
	chain, err := newObfChain("<b 0xf6ab><c><t><r 8>")
	if err != nil {
		t.Fatalf("newObfChain failed: %v", err)
	}
	if chain == nil {
		t.Fatalf("expected non-nil obfChain")
	}
	buf1 := make([]byte, chain.ObfuscatedLen(0))
	buf2 := make([]byte, chain.ObfuscatedLen(0))
	chain.Obfuscate(buf1, nil)
	chain.Obfuscate(buf2, nil)

	if !bytes.Equal(buf1[:2], []byte{0xf6, 0xab}) {
		t.Fatalf("unexpected magic prefix: %x", buf1[:2])
	}
	c1 := binary.BigEndian.Uint32(buf1[2:6])
	c2 := binary.BigEndian.Uint32(buf2[2:6])
	if c1 != 1 || c2 != 2 {
		t.Fatalf("expected counter 1 -> 2, got %d -> %d", c1, c2)
	}
}

func TestAWGHeaderProtectionShortSalt(t *testing.T) {
	d := &Device{}
	if err := d.headerProtection.key.FromHex("00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"); err != nil {
		t.Fatalf("header key parse failed: %v", err)
	}
	if cip, err := d.HeaderProtectionCipher([]byte{1, 2, 3, 4}); err == nil || cip != nil {
		t.Fatalf("expected error for short header protection salt")
	}
	validSalt := make([]byte, HeaderCipherNonceSize)
	cip, err := d.HeaderProtectionCipher(validSalt)
	if err != nil || cip == nil {
		t.Fatalf("expected valid cipher, err=%v", err)
	}
}
