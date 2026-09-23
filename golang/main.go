// Packaeg main
// nat-traversal => think of it as creating hole in the machine's firewall so that inbound traffic can pass through
package main

import (
	"app/mongoose/example"
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile)
	example.SDKV2()
	//example.LIVESDKV1()
}
