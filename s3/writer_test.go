package s3

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriter_WriteAt(t *testing.T) {
	writer := NewWriter()
	_, _ = writer.WriteAt([]byte{0x2}, 1)
	_, _ = writer.WriteAt([]byte{0x1}, 0)
	_, _ = writer.WriteAt([]byte{0x3}, 2)
	assert.Equal(t, []byte{0x1, 0x02, 0x3}, writer.Buffer)
}
