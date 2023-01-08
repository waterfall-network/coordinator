// This binary is a simple rest API endpoint to calculate
// the ENR value of a node given its private key,ip address and port.
package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"net"

	"github.com/libp2p/go-libp2p-core/crypto"
	log "github.com/sirupsen/logrus"
	"github.com/waterfall-foundation/coordinator/io/file"
	_ "github.com/waterfall-foundation/coordinator/runtime/maxprocs"
	"gitlab.waterfall.network/waterfall/protocol/gwat/p2p/enode"
	"gitlab.waterfall.network/waterfall/protocol/gwat/p2p/enr"
)

var (
	privateKey = flag.String("private", "", "Hex encoded Private key to use for calculation of ENR")
	udpPort    = flag.Int("udp-port", 0, "UDP Port to use for calculation of ENR")
	tcpPort    = flag.Int("tcp-port", 0, "TCP Port to use for calculation of ENR")
	ipAddr     = flag.String("ipAddress", "", "IP to use in calculation of ENR")
	outfile    = flag.String("out", "", "Filepath to write ENR")
)

func main() {
	flag.Parse()

	if *privateKey == "" {
		log.Fatal("No private key given")
	}
	dst, err := hex.DecodeString(*privateKey)
	if err != nil {
		panic(err)
	}
	unmarshalledKey, err := crypto.UnmarshalSecp256k1PrivateKey(dst)
	if err != nil {
		panic(err)
	}
	ecdsaPrivKey := (*ecdsa.PrivateKey)(unmarshalledKey.(*crypto.Secp256k1PrivateKey))

	if net.ParseIP(*ipAddr).To4() == nil {
		log.Fatalf("Invalid ipv4 address given: %v\n", err)
	}

	if *udpPort == 0 {
		log.Fatalf("Invalid udp port given: %v\n", err)
		return
	}

	db, err := enode.OpenDB("")
	if err != nil {
		log.Fatalf("Could not open node's peer database: %v\n", err)
		return
	}
	defer db.Close()

	localNode := enode.NewLocalNode(db, ecdsaPrivKey)
	ipEntry := enr.IP(net.ParseIP(*ipAddr))
	udpEntry := enr.UDP(*udpPort)
	localNode.Set(ipEntry)
	localNode.Set(udpEntry)
	if *tcpPort != 0 {
		tcpEntry := enr.TCP(*tcpPort)
		localNode.Set(tcpEntry)
	}
	log.Info(localNode.Node().String())

	if *outfile != "" {
		err := file.WriteFile(*outfile, []byte(localNode.Node().String()))
		if err != nil {
			panic(err)
		}
		log.Infof("Wrote to %s", *outfile)
	}
}
