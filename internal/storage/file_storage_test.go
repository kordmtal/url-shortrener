package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRepository(t *testing.T) {
	tempFile, err := os.CreateTemp("", "storage_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	rep, err := NewFileRepository(tempFile.Name())
	require.NoError(t, err)
	defer rep.Close()

	url, err := rep.Get("missing_key")
	require.NoError(t, err)
	assert.Empty(t, url)

	url, found := rep.FindByURL("missing_url")
	require.False(t, found)
	assert.Empty(t, url)

	rep.Set("testURL", "testKey")

	url, err = rep.Get("testKey")
	require.NoError(t, err)
	assert.Equal(t, "testURL", url)

	url, found = rep.FindByURL("testURL")
	require.True(t, found)
	assert.Equal(t, "testKey", url)

	rep2, err := NewFileRepository(tempFile.Name())
	require.NoError(t, err)
	defer rep2.Close()

	url, err = rep2.Get("testKey")
	require.NoError(t, err)
	assert.Equal(t, "testURL", url)

	url, found = rep2.FindByURL("testURL")
	require.True(t, found)
	assert.Equal(t, "testKey", url)
}
