package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"log"
	"time"

	pb "github.com/LabZion/HEaaS/fhe"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name used to verify the hostname returned by the TLS handshake")
)

type KeyPair struct {
	PublicKey []byte
	SecretKey []byte
}

// fetchPublicKey store a pair of fhe keys
func fetchPublicKey(client pb.FHEClient, key string) KeyPair {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	keyPair, err := client.FetchPublicKey(ctx, &pb.FetchPublicKeyRequest{
		Key: key,
	})
	if err != nil {
		log.Fatalf("%v.StoreKey(_) = _, %v: ", client, err)
	}
	return KeyPair{
		PublicKey: keyPair.PublicKey,
		SecretKey: keyPair.SecretKey,
	}
}

// fetchPublicKeyBySHA256 store a pair of fhe keys
func fetchPublicKeyBySHA256(client pb.FHEClient, key string) KeyPair {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	keyPair, err := client.FetchPublicKeyBySHA256(ctx, &pb.FetchPublicKeyRequest{
		Key: key,
	})
	if err != nil {
		log.Fatalf("%v.StoreKey(_) = _, %v: ", client, err)
	}
	return KeyPair{
		PublicKey: keyPair.PublicKey,
		SecretKey: keyPair.SecretKey,
	}
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	if *tls {
		if *caFile == "" {
			*caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewFHEClient(conn)

	keyPair := fetchPublicKey(client, "fan@torchz.net")
	pk_sha256 := sha256.Sum256(keyPair.PublicKey)
	log.Printf("public key sha256: %x", pk_sha256)
	keyPairBySHA256 := fetchPublicKeyBySHA256(client, hex.EncodeToString(pk_sha256[:]))
	if len(keyPair.SecretKey) != 0 {
		log.Fatalf("length of keyPair.SecretKey != 0, %d", len(keyPair.SecretKey))
	}
	if !bytes.Equal(keyPair.PublicKey, keyPairBySHA256.PublicKey) {
		log.Fatalf("keyPair.PublicKey != keyPairBySHA256.PublicKey")
	}
}
