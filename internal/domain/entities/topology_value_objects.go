package entities

import "time"

// NodeMetadata represents detailed metadata for a service node
type NodeMetadata struct {
	NodeID       string
	Uptime       time.Duration
	HealthScore  float64 // 0.0 to 100.0
	RequestRate  float64 // Requests per second
	ErrorRate    float64 // Errors per second
	Latency      time.Duration
	CPUUsage     float64 // Percentage
	MemoryUsage  float64 // Percentage
	ActiveConns  int64
	TotalReqs    int64
	CustomFields map[string]interface{}
}

// NewNodeMetadata creates a new node metadata instance
func NewNodeMetadata(nodeID string) *NodeMetadata {
	return &NodeMetadata{
		NodeID:       nodeID,
		HealthScore:  100.0,
		CustomFields: make(map[string]interface{}),
	}
}

// UpdateHealthScore updates the health score based on metrics
func (m *NodeMetadata) UpdateHealthScore(score float64) {
	if score < 0 {
		m.HealthScore = 0
	} else if score > 100 {
		m.HealthScore = 100
	} else {
		m.HealthScore = score
	}
}

// AddCustomField adds a custom field to the metadata
func (m *NodeMetadata) AddCustomField(key string, value interface{}) {
	if m.CustomFields == nil {
		m.CustomFields = make(map[string]interface{})
	}
	m.CustomFields[key] = value
}

// EdgeMetadata represents detailed metadata for a service connection
type EdgeMetadata struct {
	EdgeID         string
	Protocol       string
	Throughput     float64 // Messages per second
	ErrorRate      float64 // Errors per second
	AvgLatency     time.Duration
	P99Latency     time.Duration
	TotalMessages  int64
	TotalErrors    int64
	LastActiveTime time.Time
	CustomFields   map[string]interface{}
}

// NewEdgeMetadata creates a new edge metadata instance
func NewEdgeMetadata(edgeID string) *EdgeMetadata {
	return &EdgeMetadata{
		EdgeID:         edgeID,
		LastActiveTime: time.Now(),
		CustomFields:   make(map[string]interface{}),
	}
}

// UpdateThroughput updates the throughput metric
func (m *EdgeMetadata) UpdateThroughput(throughput float64) {
	m.Throughput = throughput
	m.LastActiveTime = time.Now()
}

// AddCustomField adds a custom field to the metadata
func (m *EdgeMetadata) AddCustomField(key string, value interface{}) {
	if m.CustomFields == nil {
		m.CustomFields = make(map[string]interface{})
	}
	m.CustomFields[key] = value
}

// TopologyFilters represents filters for querying topology
type TopologyFilters struct {
	ServiceTypes []string
	NodeStatuses []NodeStatus
	EdgeStatuses []EdgeStatus
	LabelFilters map[string]string
	IncludeNodes bool
	IncludeEdges bool
}

// NewTopologyFilters creates a new topology filters instance
func NewTopologyFilters() *TopologyFilters {
	return &TopologyFilters{
		ServiceTypes: make([]string, 0),
		NodeStatuses: make([]NodeStatus, 0),
		EdgeStatuses: make([]EdgeStatus, 0),
		LabelFilters: make(map[string]string),
		IncludeNodes: true,
		IncludeEdges: true,
	}
}

// AddServiceType adds a service type filter
func (f *TopologyFilters) AddServiceType(serviceType string) {
	f.ServiceTypes = append(f.ServiceTypes, serviceType)
}

// AddNodeStatus adds a node status filter
func (f *TopologyFilters) AddNodeStatus(status NodeStatus) {
	f.NodeStatuses = append(f.NodeStatuses, status)
}

// AddEdgeStatus adds an edge status filter
func (f *TopologyFilters) AddEdgeStatus(status EdgeStatus) {
	f.EdgeStatuses = append(f.EdgeStatuses, status)
}

// AddLabelFilter adds a label filter
func (f *TopologyFilters) AddLabelFilter(key, value string) {
	if f.LabelFilters == nil {
		f.LabelFilters = make(map[string]string)
	}
	f.LabelFilters[key] = value
}

// MatchesNode returns true if the node matches the filters
func (f *TopologyFilters) MatchesNode(node *ServiceNode) bool {
	// Check service type filter
	if len(f.ServiceTypes) > 0 {
		found := false
		for _, st := range f.ServiceTypes {
			if node.ServiceType == st {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check node status filter
	if len(f.NodeStatuses) > 0 {
		found := false
		for _, status := range f.NodeStatuses {
			if node.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check label filters
	for key, value := range f.LabelFilters {
		if nodeValue, exists := node.Labels[key]; !exists || nodeValue != value {
			return false
		}
	}

	return true
}

// MatchesConnection returns true if the connection matches the filters
func (f *TopologyFilters) MatchesConnection(conn *ServiceConnection) bool {
	// Check edge status filter
	if len(f.EdgeStatuses) > 0 {
		found := false
		for _, status := range f.EdgeStatuses {
			if conn.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check label filters
	for key, value := range f.LabelFilters {
		if connValue, exists := conn.Labels[key]; !exists || connValue != value {
			return false
		}
	}

	return true
}
