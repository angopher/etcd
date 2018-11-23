package influx

import (
    "context"
	"net/http"
	"net"
	"bytes"
	"time"
	"fmt"
    "go.etcd.io/etcd/raft"
    "go.etcd.io/etcd/raft/raftpb"
)

type RaftServer struct {
	val string
    Node raft.Node
    raftConfig *raft.Config
    Storage raft.Storage
    Done chan struct{}
}

func (s *RaftServer) StartNode(peers []raft.Peer) {
    s.Node = raft.StartNode(s.raftConfig, peers)
}

func (s *RaftServer) Run() {
    for {
        select {
        case rd := <-s.Node.Ready():
            s.saveToStorage(rd.HardState, rd.Entries, rd.Snapshot)
            s.send(rd.Messages)
            if !raft.IsEmptySnap(rd.Snapshot) {
                s.processSnapshot(rd.Snapshot)
            }
            for _, entry := range rd.CommittedEntries {
                s.process(entry)
                if entry.Type == raftpb.EntryConfChange {
                    var cc raftpb.ConfChange
                    cc.Unmarshal(entry.Data)
                    s.Node.ApplyConfChange(cc)
                }
            }
            s.Node.Advance()
        case <-s.Done:
            return
        }
    }
}

func (s *RaftServer) saveToStorage(h raftpb.HardState, es []raftpb.Entry, sn raftpb.Snapshot) {
}

func (s *RaftServer) send(messages []raftpb.Message) {
	for _, msg := range messages {
		data, err := msg.Marshal()
		if err != nil {
			panic(err)
		}
		err = Request("http://127.0.0.1:8888/message", data)
		if err != nil {
			panic(err)
		}
	}
}

func (s *RaftServer) processSnapshot(sn raftpb.Snapshot) {
}

func (s *RaftServer) process(entry raftpb.Entry) {
	fmt.Println("process set val:", string(entry.Data))
	s.val = string(entry.Data)
}

func (s *RaftServer) recvRaftRPC(n raft.Node, ctx context.Context, m raftpb.Message) error {
    return n.Step(ctx, m)
}

func (s *RaftServer) Propose(n raft.Node, ctx context.Context, data[]byte) error {
    return n.Propose(ctx, data)
}

func (s *RaftServer) ProposeConfChange(n raft.Node, ctx context.Context, cc raftpb.ConfChange) error {
    return n.ProposeConfChange(ctx, cc)
}


func Request(url string, data []byte) error {
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "close")

	client := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(10 * time.Second) //TODO: timeout from config
				c, err := net.DialTimeout(netw, addr, time.Second)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}


