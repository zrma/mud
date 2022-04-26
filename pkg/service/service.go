package service

import (
	"context"

	"google.golang.org/protobuf/encoding/protojson"

	"mud/pb"
)

type Publisher interface {
	Publish(body []byte) error
}

type Service struct {
	pb.UnimplementedMudServer
	Publisher Publisher
}

func (s Service) Move(_ context.Context, req *pb.MoveRequest) (*pb.MoveReply, error) {
	if err := s.createTask(req); err != nil {
		return &pb.MoveReply{
			Player: req.GetPlayer(),
			Ok:     false,
			Err:    err.Error(),
		}, nil
	}

	return &pb.MoveReply{
		Player: req.GetPlayer(),
		Ok:     true,
	}, nil
}

func (s Service) createTask(req *pb.MoveRequest) error {
	body, err := protojson.Marshal(req)
	if err != nil {
		return err
	}
	return s.Publisher.Publish(body)
}
