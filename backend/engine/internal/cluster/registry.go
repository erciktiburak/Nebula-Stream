package cluster

type NodeStatus struct {
  ID string
  LastHeartbeatUnix int64
}

type Mesh struct {
  Nodes map[string]NodeStatus
}
