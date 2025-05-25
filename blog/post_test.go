package blog

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var rawPostMd = []byte(`
title: foobar
public: true
date: 2025-02-01T14:42:55+01:00

---

hello **world**
`)

func TestParsePost(t *testing.T) {
	expDate, err := time.Parse(time.RFC3339, "2025-02-01T14:42:55+01:00")
	if err != nil {
		panic(err)
	}

	post, err := ParsePost(bytes.NewReader(rawPostMd))

	assert.Nil(t, err, "should not err")
	assert.Equal(t, "", post.Slug)
	assert.Equal(t, "foobar", post.Title)
	assert.True(t, post.Public)
	assert.True(t, post.Date.Equal(expDate))
	assert.Equal(t, "<p>hello <strong>world</strong></p>", string(post.Body))
}
