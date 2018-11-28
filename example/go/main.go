package main

import (
	"context"
	"github.com/apptreesoftware/protoc-gen-twirp_dart/example/go/config/model"
	"github.com/apptreesoftware/protoc-gen-twirp_dart/example/go/config/service"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/twitchtv/twirp"
)

type randomHaberdasher struct{}

func (h *randomHaberdasher) BuyHat(ctx context.Context, hat *model.Hat) (*model.Hat, error) {
	return hat, nil
}

func (h *randomHaberdasher) MakeHat(ctx context.Context, size *model.Size) (*model.Hat, error) {
	if int(size.Inches) <= 0 {
		return nil, twirp.InvalidArgumentError("Inches", "must be a positive number greater than zero")
	}

	ts, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil, err
	}

	return &model.Hat{
		Size:  size.Inches,
		Color: []string{"white", "black", "brown", "red", "blue"}[rand.Intn(4)],
		Name:  []string{"bowler", "baseball cap", "top hat", "derby"}[rand.Intn(3)],
		AvailableSizes: []*model.Size{
			{Inches: 10},
			{Inches: 20},
		},
		Roles: []int32{
			1,
			2,
			3,
		},
		CreatedOn: ts,
	}, nil
}

func main() {
	server := config_service.NewHaberdasherServer(&randomHaberdasher{}, nil)
	log.Fatal(http.ListenAndServe(":9000", server))
}
