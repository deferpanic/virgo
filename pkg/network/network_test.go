package network

import (
	"fmt"
	"testing"
)

func TestGenerateMAC(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println((Network{}).generateMAC())
	}
}
