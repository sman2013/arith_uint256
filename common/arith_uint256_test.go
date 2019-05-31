package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestGetCompact(t *testing.T) {
	buf := FromHex("0x1e00ffff")
	hash := ToHash("0x000000ffff000000000000000000000000000000000000000000000000000000")
	a := new(ArithUint256)
	a.FromHash(hash)
	u32 := a.GetCompact()
	buf2 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf2, u32)
	fmt.Printf("u32: %x\r\n", buf2)
	if bytes.Compare(buf2, buf) != 0 {
		panic("failed 1")
	}
	buf = a.GetBytes()
	if bytes.Compare(hash.Bytes(), buf) != 0 {
		panic("failed 2")
	}

	lHash := ToHash("0x00000000efff0000000000000000000000000000000000000000000000000000")
	b := new(ArithUint256)
	b.FromHash(lHash)
	if a.Cmp(b) <= 0 {
		panic("failed 3")
	}
	fmt.Printf("4: %x\r\n", buf)
}
