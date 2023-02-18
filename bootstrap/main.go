package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	ldb "github.com/ipfs/go-ds-leveldb"
	libp2p "github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	dhtPath = "store/dht-store/"
	keyPath = "store/host-key/"
)

// Run bootstrap Peer
func main() {
	portFlag := flag.Int("port", 1234, "port number")
	listenHost := flag.String("host", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.Parse()

	// Context
	ctx := context.Background()

	// Get or create a private key
	PathExistsOrCreates(keyPath)
	priv, err := GetHostPrivKey(keyPath)
	if err != nil {
		fmt.Print("GetHostPrivKey error")
		panic(err)
	}

	// New Host
	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", *listenHost, *portFlag))
	host, err := libp2p.New(
		// listen addresses
		libp2p.ListenAddrs(listen),
		// use private key to create host
		libp2p.Identity(priv),
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("Host: ", host.ID())

	// New DHT
	PathExistsOrCreates(dhtPath)
	mydb, err := ldb.NewDatastore(dhtPath, nil)
	if err != nil {
		panic(err)
	}

	_, err = kaddht.New(ctx, host, kaddht.Datastore(mydb))
	if err != nil {
		panic(err)
	}

	fmt.Println("")
	fmt.Printf("[*] Your Bootstrap ID Is: /ip4/%s/tcp/%v/p2p/%s\n", *listenHost, *portFlag, host.ID().Pretty())
	fmt.Println("")
}

func PathExistsOrCreates(path string) {
	_, err := os.Stat(path)

	var exist = false
	if err == nil {
		exist = true
	}
	if os.IsNotExist(err) {
		exist = false
	}

	if exist {
		//fmt.Printf("has dir![%v]\n", path)
	} else {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
		}
	}
}

func GetHostPrivKey(path string) (crypto.PrivKey, error) {
	var priv crypto.PrivKey
	// Read private key file
	tmpFile, err := os.ReadFile(path + "private.key")
	if err != nil {
		fmt.Println("Read private key file error: ", err.Error())
	} else {
		priv, err = crypto.UnmarshalPrivateKey(tmpFile)
		if err != nil {
			fmt.Println("Unmarshal private key error: ", err.Error())
		} else {
			return priv, nil
		}
	}

	// If read old privKey faild, create a new random privKey
	priv, _, err = crypto.GenerateKeyPair(crypto.ECDSA, 256)
	if err != nil {
		fmt.Println("Generate private key error: ", err.Error())
		return priv, errors.New("generate private key error")
	}

	// Store privKey to kayPath file
	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		fmt.Println("Marshal private key error: ", err.Error())
		return priv, errors.New("marshal private key error")
	}

	err = os.WriteFile(path+"private.key", privBytes, 0664)
	if err != nil {
		fmt.Println("Write private key error: ", err.Error())
		return priv, errors.New("write private key error")
	}

	return priv, err
}
