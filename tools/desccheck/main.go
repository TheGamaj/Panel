package main

import (
	"fmt"
	"os"

	nodev1 "github.com/TheGamaj/Panel/internal/proto/node/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	// Accessing the file descriptor triggers full validation of the raw bytes.
	_ = nodev1.File_gamaj_node_v1_node_proto
	// Round-trip a message to prove wire format works.
	hello := &nodev1.HelloRequest{MasterId: "test", MasterVersion: "is.0.0.1"}
	b, err := proto.Marshal(hello)
	if err != nil {
		fmt.Println("FAIL marshal:", err)
		os.Exit(1)
	}
	back := &nodev1.HelloRequest{}
	if err := proto.Unmarshal(b, back); err != nil {
		fmt.Println("FAIL unmarshal:", err)
		os.Exit(1)
	}
	if back.MasterId != "test" {
		fmt.Println("FAIL roundtrip")
		os.Exit(1)
	}
	fmt.Println("PANEL DESCRIPTOR OK — proto runtime accepts the regenerated descriptor")
}
