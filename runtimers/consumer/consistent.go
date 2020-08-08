package consumer

import (
	"hash/crc32"
	"sort"
	"strconv"

	"github.com/Shopify/sarama"
)

// ConsistentHash is a sarama.BalanceStrategy implementation with the consistent hashing algorithm.
type ConsistentHash int

type hashringNode struct {
	name         string
	hash         uint32
	partitionIdx uint32
}

type hashring []*hashringNode

func (l hashring) Len() int {
	return len(l)
}

func (l hashring) Less(i, j int) bool {
	n1 := l[i]
	n2 := l[j]

	if n1.partitionIdx < n2.partitionIdx {
		return true
	}
	if n1.partitionIdx > n2.partitionIdx {
		return false
	}
	return n1.hash < n2.hash
}

func (l hashring) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (c ConsistentHash) Name() string {
	return "ConsistentHash"
}

// Plan to consistently hash partitions.
func (c ConsistentHash) Plan(members map[string]sarama.ConsumerGroupMemberMetadata, topics map[string][]int32) (sarama.BalanceStrategyPlan, error) {
	plan := make(sarama.BalanceStrategyPlan)
	membersList := make([]string, len(members))
	i := 0
	for member := range members {
		membersList[i] = member
		i++
	}

	for topic, partitions := range topics {
		assignment := consistentHashTopic(membersList, partitions)
		for _, member := range membersList {
			if _, ok := plan[member]; !ok {
				plan[member] = make(map[string][]int32)
			}
			plan[member][topic] = append(plan[member][topic], assignment[member]...)
		}
	}

	return plan, nil
}

// Given a list of members and partitions, returns a map from member to list of paritions.
func consistentHashTopic(members []string, partitions []int32) map[string][]int32 {
	ring := hashring{}
	nHashes := 2 // TODO: make configurable
	for _, name := range members {
		for i := 0; i < nHashes; i++ {
			hash := crc32.ChecksumIEEE([]byte(name + strconv.Itoa(i)))
			ring = append(ring, &hashringNode{
				name:         name,
				hash:         hash,
				partitionIdx: hash % uint32(len(partitions)),
			})
		}
	}

	sort.Sort(ring)
	plan := make(map[string][]int32)

	for a, b := 0, 1; b < len(ring); a, b = a+1, b+1 {
		elA := ring[a]
		elB := ring[b]

		plan[elA.name] = append(plan[elA.name], partitions[elA.partitionIdx:elB.partitionIdx]...)
	}

	//circle around
	last := ring[len(ring)-1]
	first := ring[0]
	plan[last.name] = append(plan[last.name], partitions[last.partitionIdx:]...)
	plan[last.name] = append(plan[last.name], partitions[:first.partitionIdx]...)

	return plan
}

func (ConsistentHash) AssignmentData(memberID string, topics map[string][]int32, generationID int32) ([]byte, error) {
	return []byte{}, nil
}
