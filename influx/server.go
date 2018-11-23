package influx

import (
    "context"
    "go.etcd.io/etcd/raft"
    "go.etcd.io/etcd/raft/raftpb"
)

type RaftServer struct {
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
}

func (s *RaftServer) processSnapshot(sn raftpb.Snapshot) {
}

func (s *RaftServer) process(entry raftpb.Entry) {
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

