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

func (s *server) showTable() {
	fmt.Println("# Querying")
	rows, err := s.db.Query("SELECT * FROM television")
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

	defer rows.Close()

	return &television.ListTelevisionsResp{Televisions: tvs}, nil
}

func (s *server) UpdateTelevision(ctx context.Context, req *television.Television) (*television.Television, error) {
	//Doit update television.status ===> Online/Offline
	_, err := s.db.ExecContext(ctx, "update television set id=$1, name=$2, ip=$3, status=$4, composition_id=$5 where id=$6",
		req.ID, req.Name, req.IP, req.Status, req.CompositionID, req.ID)
	if err != nil {
		log.Fatal(err)
	}
	tv := television.Television{}
	row := s.db.QueryRowContext(ctx, "SELECT id, name, ip, status, composition_id FROM television WHERE id=$1", req.ID)

	if err := row.Scan(&tv.ID, &tv.Name, &tv.IP, &tv.Status, &tv.CompositionID); err != nil {
		return nil, grpc.Errorf(codes.Internal, "could not scan %v", err)
	}

	return &tv, nil
}

func (s *server) DeleteTelevision(ctx context.Context, req *television.DeleteTelevisionReq) (*empty.Empty, error) {
	_, err := s.db.ExecContext(ctx, "delete from television where id=$1", req.Id)
	if err != nil {
		log.Fatal(err)
	}

	// if err := row.Scan(&tv.ID, &tv.Name, &tv.IP, &tv.Status, &tv.CompositionID); err != nil {
	// 	return nil, grpc.Errorf(codes.Internal, "could not scan %v", err)
	// }

	return &empty.Empty{}, nil
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

	// showTable()
	srv.showTable()
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

	// DeleteTevision()
	_, err = srv.DeleteTelevision(context.Background(), &television.DeleteTelevisionReq{Id: "00000000002"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n")

	// UpdateTelevision()
	_, err = srv.UpdateTelevision(context.Background(),
		&television.Television{ID: "00000000001", Name: "Bite", IP: "192.168.0.1", Status: 4, CompositionID: "5"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n")

	// showTable()
	srv.showTable()
	fmt.Printf("\n")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
