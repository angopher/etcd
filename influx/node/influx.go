package main

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"flag"
	"context"
    "go.etcd.io/etcd/raft"
    "go.etcd.io/etcd/influx"
    "go.etcd.io/etcd/raft/raftpb"
)

func main() {
	selfID := flag.Int("id", 0, "self id")
	id2 := flag.Int("id2", 0, "peer node")
	id3 := flag.Int("id3", 0, "peer node")
	flag.Parse()

	fmt.Println(*selfID, *id2, *id3)

    storage := raft.NewMemoryStorage()
    c := &raft.Config{
        ID:              uint64(*selfID),
        ElectionTick:    10,
        HeartbeatTick:   1,
        Storage:         storage,
        MaxSizePerMsg:   4096,
        MaxInflightMsgs: 256,
    }
	peers := []raft.Peer{{ID: uint64(*selfID)}, {ID: uint64(*id2)}, {ID: uint64(*id3)}}

    srv := &influx.RaftServer{
        RaftConfig: c,
        Storage: storage,
		PeersAddr: make(map[uint64]string),
        Done: make(chan struct{}),
    }

    srv.StartNode(peers)
    go srv.Run()

    http.HandleFunc("/val", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(srv.Val)
	})
    http.HandleFunc("/set_val", func(w http.ResponseWriter, r *http.Request) {
		val := r.FormValue("val")
		fmt.Println("recv val:", val)
		srv.Propose(srv.Node, context.Background(), []byte(val))
		fmt.Println("propose done")
		w.Write([]byte("ok"))
	})
    http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("recv message")
		defer w.Write([]byte("ok"))

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		var msg raftpb.Message
		err = msg.Unmarshal(data)
		if err != nil {
			panic(err)
		}
		srv.RecvRaftRPC(srv.Node, context.Background(), msg)
	})

	fmt.Println("listen:", influx.NodeAddr(uint64(*selfID)))
	err := http.ListenAndServe(influx.NodeAddr(uint64(*selfID)), nil)
    if err != nil {
        fmt.Printf("msg=ListenAndServe failed,err=%s\n", err.Error())
    }
}


