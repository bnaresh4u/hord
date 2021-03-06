package app

import (
	"context"
	"fmt"
	"github.com/madflojo/hord/databases"
	pb "github.com/madflojo/hord/proto/client"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"time"
)

// server is used to implement the client protobuf server interface
type grpcServer struct{}

// grpcListener will start the grpc server listening on the defined grpc port
func grpcListener() error {
	lis, err := net.Listen("tcp", Config.GRPCPort)
	if err != nil {
		return err
	}
	srv := grpc.NewServer()
	pb.RegisterHordServer(srv, &grpcServer{})
	err = srv.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}

// Get will retrieve requested information from the datastore and return it
func (s *grpcServer) Get(ctx context.Context, msg *pb.GetRequest) (*pb.GetResponse, error) {
	// Define reply message
	r := &pb.GetResponse{
		Status: &pb.Status{
			Code:        0,
			Description: "Success",
		},
	}

	// Check key length
	if len(msg.Key) == 0 {
		log.WithFields(logrus.Fields{"key": msg.Key}).Tracef("Key %s is not defined", msg.Key)
		r.Status.Code = 4
		r.Status.Description = "Key must be defined"
		return r, fmt.Errorf("Key %s is not defined", msg.Key)
	}

	// Fetch data using key
	d, err := db.Get(msg.Key)
	if err != nil {
		log.WithFields(logrus.Fields{"key": msg.Key, "error": err}).Tracef("Failed to fetch data for key - %s", err)
		r.Status.Code = 5
		r.Status.Description = "Error fetching data from datastore"
		return r, fmt.Errorf("Error fetching data from datastore - %s", err)
	}

	// Return data to client
	r.Key = msg.Key
	r.Data = d.Data
	r.LastUpdated = d.LastUpdated
	return r, nil
}

// Set will take the supplied data and store it within the datastore returning success or failure
func (s *grpcServer) Set(ctx context.Context, msg *pb.SetRequest) (*pb.SetResponse, error) {
	// Define reply message
	r := &pb.SetResponse{
		Status: &pb.Status{
			Code:        0,
			Description: "Success",
		},
	}

	// Check key length
	if len(msg.Key) == 0 {
		log.WithFields(logrus.Fields{"key": msg.Key}).Tracef("Key %s is not defined", msg.Key)
		r.Status.Code = 4
		r.Status.Description = "Key must be defined"
		return r, fmt.Errorf("Key %s is not defined", msg.Key)
	}

	// Create data item for insertion
	d := &databases.Data{}
	d.Data = msg.Data
	d.LastUpdated = time.Now().UnixNano()

	// Insert data into datastore
	// TODO: Handle TTL settings
	err := db.Set(msg.Key, d)
	if err != nil {
		log.WithFields(logrus.Fields{"key": msg.Key, "error": err}).Tracef("Failed to store data for key - %s", err)
		r.Status.Code = 5
		r.Status.Description = "Error storing data within datastore"
		return r, fmt.Errorf("Error storing data within datastore - %s", err)
	}

	r.Key = msg.Key
	return r, nil
}

// Delete will remove the specified key from the datastore and return success or failure
func (s *grpcServer) Delete(ctx context.Context, msg *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	return &pb.DeleteResponse{}, nil
}
