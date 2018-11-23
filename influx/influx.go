package influx

import (
    "go.etcd.io/etcd/raft"
)

func main() {
    storage := raft.NewMemoryStorage()
    c := &raft.Config{
        ID:              0x01,
        ElectionTick:    10,
        HeartbeatTick:   1,
        Storage:         storage,
        MaxSizePerMsg:   4096,
        MaxInflightMsgs: 256,
    }
    peers := []raft.Peer{{ID: 0x02}, {ID: 0x03}}

    srv := &RaftServer{
        raftConfig: c,
        Storage: storage,
        Done: make(chan struct{}),
    }

    srv.StartNode(peers)
    srv.Run()
}



