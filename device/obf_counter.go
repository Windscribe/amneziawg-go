package device

import (
	"encoding/binary"
	"sync/atomic"
)

func newCounterObf(_ string) (obf, error) {
	return &counterObf{}, nil
}

type counterObf struct {
	seq atomic.Uint32
}

func (o *counterObf) Obfuscate(dst, src []byte) {
	binary.BigEndian.PutUint32(dst, o.seq.Add(1))
}

func (o *counterObf) Deobfuscate(dst, src []byte) bool {
	return len(src) >= 4
}

func (o *counterObf) ObfuscatedLen(n int) int {
	return 4
}

func (o *counterObf) DeobfuscatedLen(n int) int {
	return 0
}
