package util

import (
	"testing"

	"github.com/konpyutaika/nifikop/api/v1alpha1"
)

func TestSubtractNodes(t *testing.T) {
	sourceList := []v1alpha1.Node{
		{
			Id: 1,
		},
		{
			Id: 2,
		},
		{
			Id: 3,
		},
	}

	nodesToSubtract := []v1alpha1.Node{
		{
			Id: 3,
		},
	}

	// subtract 1 node
	if results := SubtractNodes(sourceList, nodesToSubtract); len(results) != 2 {
		t.Error("There should be two nodes remaining")
	}

	// subtract empty list
	if results := SubtractNodes(sourceList, []v1alpha1.Node{}); len(results) != 3 {
		t.Error("there should be 3 results")
	}

	// subtract all nodes
	if results := SubtractNodes(sourceList, sourceList); len(results) != 0 {
		t.Error("There should be two nodes remaining")
	}
}
