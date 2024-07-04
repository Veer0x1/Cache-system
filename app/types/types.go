package types

import (
	"sync"
)

type DBItem struct {
	Value  string
	Expiry int64
}

type StreamEntry struct {
	ID  string
	KVs map[string]string
}

type ServerState struct {
	DB      map[string]DBItem
	Streams map[string][]StreamEntry
	DBMutex sync.Mutex
	Port    int

	DBDir      string // Directory in which to store the database files
	DBFilename string // Name of the database file

	Role             string    // master | slave
	MasterReplID     string    // Replication ID of the master (own replication ID if master)
	MasterReplOffset int       // Offset of the master (0 if master)
	MasterHost       string    // Host of the master (empty if master)
	MasterPort       string    // Port of the master (empty if master)
	Replicas         []Replica // Connections to replicas (empty if slave)
	AckOffset        int       // Offset of the last acknowledged replication message (only for slaves)
	BytesSent        int       // Number of bytes sent to replicas (only for masters)
}