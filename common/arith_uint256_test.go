package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"testing"
)

func TestGetCompact(t *testing.T) {
	buf := bytes.NewBuffer(FromHex("0x1b0404cb"))
	var x uint32
	binary.Read(buf, binary.BigEndian, &x)
	fmt.Printf("%x\r\n", buf)
	fmt.Println(x)
	u64, _ := strconv.ParseUint("1b0404cb", 16, 32)
	fmt.Println(u64)
	a := new(ArithUint256)
	a.SetCompact(uint32(u64))
	u32 := a.GetCompact()
	buf2 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf2, u32)
	fmt.Printf("%x\r\n", buf2)
}
