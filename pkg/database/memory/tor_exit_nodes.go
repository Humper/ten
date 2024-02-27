package memory

import (
	"context"
	"math"
	"sort"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/humper/tor_exit_nodes/models"
	"gorm.io/gorm"
)

type torExitNodes struct {
	nodes map[string]*models.TorExitNode
	mutex sync.Mutex
}

func copyExitNode(node *models.TorExitNode) *models.TorExitNode {
	return &models.TorExitNode{
		Model:       gorm.Model{ID: node.ID},
		IP:          node.IP,
		CountryCode: node.CountryCode,
		CountryName: node.CountryName,
	}
}

func (t *torExitNodes) GetAll(ctx context.Context, country_codes []string, excludedIPs []string, pagination *models.Pagination) (*models.Pagination, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	allNodes := []*models.TorExitNode{}
	for _, node := range t.nodes {
		allNodes = append(allNodes, node)
	}
	sort.Slice(allNodes, func(i, j int) bool {
		return allNodes[i].ID < allNodes[j].ID
	})

	exclusionSet := mapset.NewSet[string]()
	exclusionSet.Append(excludedIPs...)

	countrySet := mapset.NewSet[string]()
	countrySet.Append(country_codes...)

	filteredNodes := []*models.TorExitNode{}
	for _, node := range allNodes {
		if (len(country_codes) == 0 || countrySet.Contains(node.CountryCode)) && !exclusionSet.Contains(node.IP) {
			filteredNodes = append(filteredNodes, node)
		}
	}
	data := make([]*models.TorExitNode, 0, pagination.GetLimit())

	totalRows := len(filteredNodes)

	pagination.TotalRows = int64(totalRows)
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.GetLimit())))
	pagination.TotalPages = totalPages

	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if end > totalRows {
		end = totalRows
	}

	for i := start; i < end; i++ {
		data = append(data, copyExitNode(filteredNodes[i]))
	}

	pagination.Rows = data
	return pagination, nil
}

func (t *torExitNodes) GetUniqueCountryCodes(ctx context.Context, pagination *models.Pagination) (*models.Pagination, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	countrySet := mapset.NewSet[string]()

	allNodes := []*models.TorExitNode{}
	for _, node := range t.nodes {
		if node.CountryCode != "" && !countrySet.Contains(node.CountryCode) {
			countrySet.Add(node.CountryCode)
			allNodes = append(allNodes, node)
		}
	}
	sort.Slice(allNodes, func(i, j int) bool {
		return allNodes[i].ID < allNodes[j].ID
	})

	data := make([]*models.TorExitNode, 0, pagination.GetLimit())

	totalRows := len(allNodes)

	pagination.TotalRows = int64(totalRows)
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.GetLimit())))
	pagination.TotalPages = totalPages

	start := pagination.GetOffset()
	end := start + pagination.GetLimit()
	if end > totalRows {
		end = totalRows
	}

	for i := start; i < end; i++ {
		data = append(data, copyExitNode(allNodes[i]))
	}

	pagination.Rows = data
	return pagination, nil
}

func (t *torExitNodes) DeleteAndAdd(ctx context.Context, nodes_to_delete []models.TorExitNode, nodes_to_add []*models.TorExitNode) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, node := range nodes_to_delete {
		delete(t.nodes, node.IP)
	}
	for _, node := range nodes_to_add {
		t.nodes[node.IP] = copyExitNode(node)
		t.nodes[node.IP].ID = uint(len(t.nodes))
	}
	return nil
}

func (t *torExitNodes) GetMissingCountries(ctx context.Context, batchSize int) ([]*models.TorExitNode, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	nodes := make([]*models.TorExitNode, 0, batchSize)
	for _, node := range t.nodes {
		if node.CountryName == "" {
			nodes = append(nodes, copyExitNode(node))
		}
	}
	return nodes, nil
}

func (t *torExitNodes) Update(ctx context.Context, nodes []*models.TorExitNode) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, node := range nodes {
		t.nodes[node.IP] = copyExitNode(node)
	}
	return nil
}
