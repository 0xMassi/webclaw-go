package webclaw

import (
	"reflect"
	"testing"
)

func TestClientDoesNotExposeRetiredLeadMethods(t *testing.T) {
	clientType := reflect.TypeOf((*Client)(nil))
	for _, name := range []string{"Lead", "LeadBatch", "GetLeadBatch", "WaitForLeadBatch"} {
		if _, exists := clientType.MethodByName(name); exists {
			t.Errorf("retired method %s is still exposed", name)
		}
	}
}
