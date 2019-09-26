// Copyright 2019 The vogo Authors. All rights reserved.
// author: wongoo

package gracego

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilePath(t *testing.T) {
	assert.Equal(t, ".", filepath.Dir("test"))
	assert.Equal(t, ".", filepath.Dir("."))
	assert.Equal(t, ".", filepath.Dir("./"))
	assert.Equal(t, ".", filepath.Dir("./test"))
	assert.Equal(t, "..", filepath.Dir("./../test"))
	assert.Equal(t, "../..", filepath.Dir("./../../test"))

	assert.Equal(t, "/a/b/test", filepath.Clean("/a/b/test"))
	assert.Equal(t, "/a/b/test", filepath.Clean("/a/b/test/"))
	assert.Equal(t, "/test", filepath.Clean("/a/b/../../test"))

	assert.Equal(t, "test", filepath.Clean("test"))
	assert.Equal(t, "test", filepath.Clean("./test"))
	assert.Equal(t, "../test", filepath.Clean("./../test"))
	assert.Equal(t, "../../test", filepath.Clean("./../../test"))
}
