package utills

import (
	"fmt"
	"github.com/pedroxer/booking-service/internal/config"
	proto_gen "github.com/pedroxer/booking-service/internal/proto_gen/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CreateResourceClient(config config.ResourceService) (proto_gen.ResourceServiceClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", config.Host, config.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := proto_gen.NewResourceServiceClient(conn)
	return client, nil
}
