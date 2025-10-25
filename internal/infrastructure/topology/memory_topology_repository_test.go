package topology

import (
	"context"
	"fmt"
	"testing"

	"github.com/quantfidential/trading-ecosystem/audit-correlator-go/internal/domain/entities"
)

func TestMemoryTopologyRepository_NodeOperations(t *testing.T) {
	repo := NewMemoryTopologyRepository()
	ctx := context.Background()

	// Test SaveNode
	node := entities.NewServiceNode("node-1", "risk-monitor", "risk-monitor-py", "risk-monitor-lh")
	err := repo.SaveNode(ctx, node)
	if err != nil {
		t.Fatalf("SaveNode failed: %v", err)
	}

	// Test GetNode
	retrieved, err := repo.GetNode(ctx, "node-1")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if retrieved.ID != "node-1" {
		t.Errorf("Expected node ID 'node-1', got %s", retrieved.ID)
	}

	// Test GetNodes
	nodes, err := repo.GetNodes(ctx, nil)
	if err != nil {
		t.Fatalf("GetNodes failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}

	// Test DeleteNode
	err = repo.DeleteNode(ctx, "node-1")
	if err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetNode(ctx, "node-1")
	if err == nil {
		t.Error("Expected error after deleting node")
	}
}

func TestMemoryTopologyRepository_ConnectionOperations(t *testing.T) {
	repo := NewMemoryTopologyRepository()
	ctx := context.Background()

	// Test SaveConnection
	conn := entities.NewServiceConnection("conn-1", "node-1", "node-2", entities.ConnectionTypeGRPC)
	err := repo.SaveConnection(ctx, conn)
	if err != nil {
		t.Fatalf("SaveConnection failed: %v", err)
	}

	// Test GetConnection
	retrieved, err := repo.GetConnection(ctx, "conn-1")
	if err != nil {
		t.Fatalf("GetConnection failed: %v", err)
	}
	if retrieved.ID != "conn-1" {
		t.Errorf("Expected connection ID 'conn-1', got %s", retrieved.ID)
	}

	// Test GetConnections
	connections, err := repo.GetConnections(ctx, nil)
	if err != nil {
		t.Fatalf("GetConnections failed: %v", err)
	}
	if len(connections) != 1 {
		t.Errorf("Expected 1 connection, got %d", len(connections))
	}

	// Test DeleteConnection
	err = repo.DeleteConnection(ctx, "conn-1")
	if err != nil {
		t.Fatalf("DeleteConnection failed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetConnection(ctx, "conn-1")
	if err == nil {
		t.Error("Expected error after deleting connection")
	}
}

func TestMemoryTopologyRepository_TopologySnapshot(t *testing.T) {
	repo := NewMemoryTopologyRepository()
	ctx := context.Background()

	// Add some nodes and connections
	node1 := entities.NewServiceNode("node-1", "test1", "type1", "instance1")
	node2 := entities.NewServiceNode("node-2", "test2", "type2", "instance2")
	conn := entities.NewServiceConnection("conn-1", "node-1", "node-2", entities.ConnectionTypeGRPC)

	repo.SaveNode(ctx, node1)
	repo.SaveNode(ctx, node2)
	repo.SaveConnection(ctx, conn)

	// Test GetTopologySnapshot
	snapshot, err := repo.GetTopologySnapshot(ctx)
	if err != nil {
		t.Fatalf("GetTopologySnapshot failed: %v", err)
	}

	if len(snapshot.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in snapshot, got %d", len(snapshot.Nodes))
	}

	if len(snapshot.Connections) != 1 {
		t.Errorf("Expected 1 connection in snapshot, got %d", len(snapshot.Connections))
	}

	// Test SaveTopologySnapshot
	newTopology := entities.NewNetworkTopology("snapshot-new")
	newNode := entities.NewServiceNode("node-3", "test3", "type3", "instance3")
	newTopology.AddNode(newNode)

	err = repo.SaveTopologySnapshot(ctx, newTopology)
	if err != nil {
		t.Fatalf("SaveTopologySnapshot failed: %v", err)
	}

	// Verify new snapshot
	updatedSnapshot, err := repo.GetTopologySnapshot(ctx)
	if err != nil {
		t.Fatalf("GetTopologySnapshot after save failed: %v", err)
	}

	if len(updatedSnapshot.Nodes) != 1 {
		t.Errorf("Expected 1 node after save, got %d", len(updatedSnapshot.Nodes))
	}

	if updatedSnapshot.SnapshotID != "snapshot-new" {
		t.Errorf("Expected snapshot ID 'snapshot-new', got %s", updatedSnapshot.SnapshotID)
	}
}

func TestMemoryTopologyRepository_FilteredQueries(t *testing.T) {
	repo := NewMemoryTopologyRepository()
	ctx := context.Background()

	// Add nodes with different service types
	node1 := entities.NewServiceNode("node-1", "risk1", "risk-monitor-py", "instance1")
	node2 := entities.NewServiceNode("node-2", "risk2", "risk-monitor-py", "instance2")
	node3 := entities.NewServiceNode("node-3", "exchange", "exchange-simulator-go", "instance3")

	repo.SaveNode(ctx, node1)
	repo.SaveNode(ctx, node2)
	repo.SaveNode(ctx, node3)

	// Test filtering by service type
	filters := entities.NewTopologyFilters()
	filters.AddServiceType("risk-monitor-py")

	nodes, err := repo.GetNodes(ctx, filters)
	if err != nil {
		t.Fatalf("GetNodes with filters failed: %v", err)
	}

	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes matching filter, got %d", len(nodes))
	}
}

func TestMemoryTopologyRepository_ConcurrentAccess(t *testing.T) {
	repo := NewMemoryTopologyRepository()
	ctx := context.Background()

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			node := entities.NewServiceNode(
				fmt.Sprintf("node-%d", id),
				fmt.Sprintf("test-%d", id),
				"test-type",
				fmt.Sprintf("instance-%d", id),
			)
			repo.SaveNode(ctx, node)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all nodes were saved
	nodes, err := repo.GetNodes(ctx, nil)
	if err != nil {
		t.Fatalf("GetNodes failed: %v", err)
	}

	if len(nodes) != 10 {
		t.Errorf("Expected 10 nodes after concurrent writes, got %d", len(nodes))
	}
}
