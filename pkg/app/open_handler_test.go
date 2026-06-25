package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveLazyCatFilePath_DirectPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "run.gpx")
	require.NoError(t, os.WriteFile(filePath, []byte("<gpx/>"), 0o644))

	got, err := resolveLazyCatFilePath(filePath)
	require.NoError(t, err)
	assert.Equal(t, filePath, got)
}

func TestResolveLazyCatFilePath_HomeMapping(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	uid := "alice"
	fileRest := "Sports/run.gpx"
	mapped := filepath.Join(dir, uid, fileRest)
	require.NoError(t, os.MkdirAll(filepath.Dir(mapped), 0o755))
	require.NoError(t, os.WriteFile(mapped, []byte("<gpx/>"), 0o644))

	got, err := resolveLazyCatFilePath("/home/" + uid + "/" + fileRest)
	require.Error(t, err)
	assert.Empty(t, got)

	got, err = resolveLazyCatFilePath(mapped)
	require.NoError(t, err)
	assert.Equal(t, mapped, got)
}

func TestResolveLazyCatFilePath_Empty(t *testing.T) {
	t.Parallel()

	got, err := resolveLazyCatFilePath("")
	require.Error(t, err)
	assert.Empty(t, got)
	assert.ErrorIs(t, err, ErrOpenMissingFile)
}
