package network

import pubsub "github.com/libp2p/go-libp2p-pubsub"

func SetupScoreParamsAndthresholds() (*pubsub.PeerScoreParams, *pubsub.PeerScoreThresholds) {
	// New PeerScoreParams and PeerScoreThresholds
	scoreParams := &pubsub.PeerScoreParams{}
	scorethresholds := &pubsub.PeerScoreThresholds{}

	return scoreParams, scorethresholds
}
