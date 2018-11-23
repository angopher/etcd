package influx

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"context"
    "go.etcd.io/etcd/raft"
    "go.etcd.io/etcd/raft/raftpb"
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
    go srv.Run()

    http.HandleFunc("/val", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(srv.val)
	})
    http.HandleFunc("/set_val", func(w http.ResponseWriter, r *http.Request) {
		val := r.FormValue("val")
		fmt.Println("recv val:", val)
		srv.Propose(srv.Node, context.Background(), []byte(val))
	})
    http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
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
		srv.recvRaftRPC(srv.Node, context.Background(), msg)
	})

	err := http.ListenAndServe(fmt.Sprintf(":%d", 8888), nil)
    if err != nil {
        fmt.Printf("msg=ListenAndServe failed,err=%s\n", err.Error())
    }
}


