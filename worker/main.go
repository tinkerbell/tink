package main

import (
	"log"

	"github.com/packethost/rover/client"
	pb "github.com/packethost/rover/protos/rover"
)

var rClient pb.RoverClient

func main() {
	conn, err := client.GetConnection()
	if err != nil {
		log.Fatal(err)
	}
	rClient = pb.NewRoverClient(conn)
	initializeWorker(rClient)
}
