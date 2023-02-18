package main

import (
	net "abic/network"
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

var (
	dhtPath   = "store/dht-store/"
	keyPath   = "store/host-key/"
	topicName = "abaci-libp2p-02-17-17-04"
)

func main() {
	portFlag := flag.Int("port", 1234, "port number")
	flag.Parse()
	port := fmt.Sprintf("%d", *portFlag)

	network, err := net.NewNetwork(port, dhtPath, keyPath, false)
	if err != nil {
		fmt.Println(err)
	}

	topic, sub, err := network.JoinTopic(topicName)
	if err != nil {
		fmt.Println(err)
	}

	ctx := context.Background()

	go streamConsoleTo(ctx, topic)

	printMessagesFrom(ctx, sub)
}

func streamConsoleTo(ctx context.Context, topic *pubsub.Topic) {
	reader := bufio.NewReader(os.Stdin)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if err := topic.Publish(ctx, []byte(s)); err != nil {
			fmt.Println("### Publish error:", err)
		}
	}
}

func printMessagesFrom(ctx context.Context, sub *pubsub.Subscription) {
	for {
		m, err := sub.Next(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Println(m.ReceivedFrom, ": ", string(m.Message.Data))
	}
}
