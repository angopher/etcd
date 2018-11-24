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
	Val string
    Node raft.Node
    RaftConfig *raft.Config
    Storage *raft.MemoryStorage
	PeersAddr map[uint64]string
    Done chan struct{}
}

func NodeAddr(id uint64) string {
	return fmt.Sprintf("127.0.0.1:800%d", id)
}

func (s *RaftServer) StartNode(peers []raft.Peer) {
	for _, peer := range peers {
		s.PeersAddr[peer.ID] = NodeAddr(peer.ID)
	}
	fmt.Printf("peers:%+v\n", peers)
    s.Node = raft.StartNode(s.RaftConfig, peers)
}

func (s *RaftServer) Run() {
	t := time.NewTicker(20 * time.Millisecond)
    for {
        select {
		case <-t.C:
			s.Node.Tick()
        case rd := <-s.Node.Ready():
            s.saveToStorage(rd.HardState, rd.Entries, rd.Snapshot)
            s.send(rd.Messages)
            if !raft.IsEmptySnap(rd.Snapshot) {
                s.processSnapshot(rd.Snapshot)
            }
            for _, entry := range rd.CommittedEntries {
                s.process(entry)
            }
            s.Node.Advance()
        case <-s.Done:
            return
        }
    }
}

func (s *RaftServer) saveToStorage(h raftpb.HardState, es []raftpb.Entry, sn raftpb.Snapshot) {
	s.Storage.Append(es)
	s.Storage.SetHardState(h)
	s.Storage.ApplySnapshot(sn)
}

func (s *RaftServer) send(messages []raftpb.Message) {
	fmt.Println("send msg num:", len(messages))
	for _, msg := range messages {
		data, err := msg.Marshal()
		if err != nil {
			panic(err)
		}
		err = Request(fmt.Sprintf("%s/message", NodeAddr(msg.To)), data)
		if err != nil {
			panic(err)
		}
	}
}

func (s *RaftServer) processSnapshot(sn raftpb.Snapshot) {
	fmt.Println("processSnapshot")
}

func (s *RaftServer) process(entry raftpb.Entry) {
	fmt.Println("process entry")

	if entry.Type == raftpb.EntryConfChange {
		var cc raftpb.ConfChange
		cc.Unmarshal(entry.Data)
		fmt.Printf("conf change: %+v\n", cc)
		s.Node.ApplyConfChange(cc)
	} else if len(entry.Data) == 0 {
		fmt.Println("empty entry")
	} else {
		fmt.Println("process set val:", string(entry.Data))
		s.Val = string(entry.Data)
	}

}

func (s *RaftServer) RecvRaftRPC(n raft.Node, ctx context.Context, m raftpb.Message) error {
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


