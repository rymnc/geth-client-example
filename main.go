package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kubemq-io/kubemq-go"
)

type FuncKubeMqClient func() *kubemq.EventsClient

func getKubeMqClient(ctx context.Context) FuncKubeMqClient {
	client, err := kubemq.NewEventsClient(
		ctx,
		kubemq.WithAddress("localhost", 50000),
		kubemq.WithClientId("client_id"),
		kubemq.WithTransportType(kubemq.TransportTypeGRPC),
		kubemq.WithCheckConnection(true),
	)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	return func() *kubemq.EventsClient {
		return client
	}
}

func publish(ctx context.Context, pubClient kubemq.EventsClient, cleanedBlock CleanedBlock) {
	jsonified, err := json.Marshal(cleanedBlock)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	errPub := pubClient.Send(
		ctx,
		kubemq.NewEvent().SetChannel("newBlocks").SetBody(jsonified),
	)
	if errPub != nil {
		log.Fatal().Msg(errPub.Error())
	}
}

type CleanedBlock struct {
	Hash             string `json:"hash"`
	Number           uint64 `json:"number"`
	Time             uint64 `json:"time"`
	TransactionCount int    `json:"transactionCount"`
}

func formatBlock(block *types.Block) CleanedBlock {
	cleanedBlock := CleanedBlock{
		Hash:             block.Hash().Hex(),
		Number:           block.Number().Uint64(),
		Time:             block.Time(),
		TransactionCount: len(block.Transactions()),
	}
	return cleanedBlock
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Msg("Error loading .env file")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	client, err := ethclient.Dial(os.Getenv("RPC_URL"))
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msg("Connected to Alchemy")

	pubClient := getKubeMqClient(ctx)()
	log.Info().Msg("Connected to KubeMQ")

	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msg("Subscribed to new headers")

	for {
		select {
		case err := <-sub.Err():
			log.Fatal().Msg(err.Error())
		case header := <-headers:
			log.Info().Msg("Got new block")
			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal().Msg(err.Error())
			}
			go publish(ctx, *pubClient, formatBlock(block))
		}
	}
}
