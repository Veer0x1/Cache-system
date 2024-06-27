package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type ReplicaInfo struct {
	ListeningPort string
	ReplicaIP     string
	Connection    *net.Conn
}

type ReplicaManager struct {
	ReplicasMap sync.Map
}

func NewReplicaManager() *ReplicaManager {
	return &ReplicaManager{}
}

func (rm *ReplicaManager) AddReplica(ip string, port string, conn *net.Conn) {
	key := ip + ":" + port
	rm.ReplicasMap.Store(key, ReplicaInfo{Connection: conn, ListeningPort: port, ReplicaIP: ip})
}

func (rm *ReplicaManager) RemoveReplica(ip string, port string) {
	key := ip + ":" + port
	rm.ReplicasMap.Delete(key)
}

func (rm *ReplicaManager) GetReplicaInfo(ip string, port string) (ReplicaInfo, bool) {
	key := ip + ":" + port
	value, ok := rm.ReplicasMap.Load(key)
	if ok {
		return value.(ReplicaInfo), true
	}
	return ReplicaInfo{}, false
}

func (rm *ReplicaManager) SendUpdateToReplicas(command string) {
	var replicaCount int
	rm.ReplicasMap.Range(func(_, _ interface{}) bool {
		replicaCount++
		return true
	})

	// Channel for collecting errors from goroutines
	errChan := make(chan error, replicaCount)

	rm.ReplicasMap.Range(func(key, value interface{}) bool {
		replicaInfo := value.(ReplicaInfo)
		go func(replicaInfo ReplicaInfo) {

			fmt.Print("Sending update to replica: ", replicaInfo.ReplicaIP, ":", replicaInfo.ListeningPort, "\n")
			//print connection
			fmt.Println(replicaInfo.Connection)
			fmt.Println(command)

			_, err := (*replicaInfo.Connection).Write([]byte(command))
			if err != nil {
				errChan <- err
			}
		}(replicaInfo)
		return true
	})

	for i := 0; i < replicaCount; i++ {
		err := <-errChan
		if err != nil {
			log.Printf("Failed to send update to replica: %v", err)
		}
	}
	close(errChan)
}
