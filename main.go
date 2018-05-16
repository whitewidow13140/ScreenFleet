package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc/codes"

	"github.com/golang/protobuf/ptypes/empty"
	_ "github.com/lib/pq"
	"github.com/skwair/screen-fleet-proto/television"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

const (
	dbuser     = "postgres"
	dbpassword = "apassword"
	dbname     = "tvserver"
)

// ConnectToDB is used to connect to DB :D
func ConnectToDB() {
	//	dbinfo := fmt.Sprintf("postgres://%s:%s@localhost:5432/tvserver?sslmode=disable", dbuser, dbpassword)
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		dbuser, dbpassword, dbname)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	fmt.Println("# Querying")
	rows, err := db.Query("SELECT * FROM television")
	checkErr(err)

	fmt.Printf("%20v | %15v | %15v | %10v | %15v \n\n", "id", "name", "ip", "status", "compositionID")
	for rows.Next() {
		var id string
		var name string
		var ip string
		var status int
		var compositionID string
		err = rows.Scan(&id, &name, &ip, &status, &compositionID)
		checkErr(err)
		fmt.Printf("%20v | %15v | %15v | %10v | %15v\n", id, name, ip, status, compositionID)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	db *sql.DB
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func (s *server) GetTelevision(ctx context.Context, req *television.GetTelevisionReq) (*television.Television, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, name, ip, status, composition_id FROM television WHERE id=$1", req.Id)
	tv := television.Television{}

	if err := row.Scan(&tv.ID, &tv.Name, &tv.IP, &tv.Status, &tv.CompositionID); err != nil {
		return nil, grpc.Errorf(codes.Internal, "could not scan %v", err)
	}

	return &tv, nil
}

func (s *server) ListTelevisions(ctx context.Context, req *television.ListTelevisionsReq) (*television.ListTelevisionsResp, error) {
	tvs := []*television.Television{}
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, ip, status, composition_id FROM television LIMIT $1 OFFSET $2", req.Size_, req.From)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		tv := television.Television{}
		if err := rows.Scan(&tv.ID, &tv.Name, &tv.IP, &tv.Status, &tv.CompositionID); err != nil {
			return nil, grpc.Errorf(codes.Internal, "could not scan %v", err)
		}
		tvs = append(tvs, &tv)
	}
	err = rows.Err()

	//	if err := rows.Scan(&tv.ID, &tv.Name, &tv.IP, &tv.Status, &tv.CompositionID); err != nil {
	defer rows.Close()

	return &television.ListTelevisionsResp{Televisions: tvs}, nil
}

func (s *server) UpdateTelevision(ctx context.Context) (*television.Television, error) {
	//Doit update television.status ===> Online/Offline
	return &television.Television{}, nil
}

func (s *server) DeleteTelevision(ctx context.Context, req television.DeleteTelevisionReq) empty.Empty {
	return empty.Empty{}
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	//Create DB
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		dbuser, dbpassword, dbname)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	srv := &server{
		db: db,
	}
	pb.RegisterGreeterServer(s, srv)
	// Register reflection service on gRPC server.
	reflection.Register(s)

	///////////////////////
	//// TESTS ////////////
	///////////////////////

	// ConnectToDB()
	fmt.Printf("\n")

	// GetTelevisions()
	getResp, err := srv.GetTelevision(context.Background(), &television.GetTelevisionReq{Id: "00000000001"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(getResp)
	fmt.Printf("\n")

	// ListTelevisions()
	resp, err := srv.ListTelevisions(context.Background(), &television.ListTelevisionsReq{From: 0, Size_: 5})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
	fmt.Printf("\n")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
