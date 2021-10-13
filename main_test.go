package main

import (
	"context"

	"testing"
	"time"
)

func benchmark(count int, b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pubClient := getKubeMqClient(ctx)()

	testBlock := CleanedBlock{
		Hash:             "0x0000000000000000000000000000000000000000000000000000000000000000",
		Number:           0,
		Time:             uint64(time.Now().Unix()),
		TransactionCount: 150,
	}
	for i := 0; i < count; i++ {
		publish(ctx, *pubClient, testBlock)
	}
}

func BenchmarkPublish1(b *testing.B) {
	benchmark(1, b)
}

func BenchmarkPublish100(b *testing.B) {
	benchmark(100, b)
}

func BenchmarkPublish1000(b *testing.B) {
	benchmark(1000, b)
}
