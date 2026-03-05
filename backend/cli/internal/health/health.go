package health

import "fmt"

type NodeReport struct {
	NodeID string
	Status string
}

func NodeHealthSummary(nodeID string) NodeReport {
	id := nodeID
	if id == "" {
		id = "local"
	}

	return NodeReport{
		NodeID: id,
		Status: "ok",
	}
}

func Render(report NodeReport) string {
	return fmt.Sprintf("node=%s status=%s", report.NodeID, report.Status)
}
