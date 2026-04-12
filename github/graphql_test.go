package github

import (
	"encoding/json"
	"time"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixture mirrors the shape returned by GitHub's GraphQL API after Phase 1
// merged comments/reviews/reviewThreads into the single FetchPR query.
const mergedPRFixture = `{
  "number": 7,
  "title": "feat: thing",
  "author": {"login": "alice", "avatarUrl": "https://a.example/alice.png"},
  "headRefName": "feat/thing",
  "baseRefName": "main",
  "state": "OPEN",
  "isDraft": false,
  "reviewDecision": "APPROVED",
  "mergeable": "MERGEABLE",
  "url": "https://github.com/org/repo/pull/7",
  "commits": {"nodes": []},
  "comments": {
    "nodes": [
      {
        "databaseId": 100,
        "body": "LGTM",
        "author": {"login": "bob", "avatarUrl": "https://a.example/bob.png"},
        "url": "https://github.com/org/repo/pull/7#issuecomment-100",
        "createdAt": "2026-01-01T12:00:00Z"
      }
    ]
  },
  "reviews": {
    "nodes": [
      {
        "databaseId": 200,
        "body": "top-level review body",
        "author": {"login": "carol"},
        "url": "https://github.com/org/repo/pull/7#pullrequestreview-200",
        "createdAt": "2026-01-01T13:00:00Z"
      },
      {
        "databaseId": 201,
        "body": "",
        "author": {"login": "dave"},
        "url": "https://github.com/org/repo/pull/7#pullrequestreview-201",
        "createdAt": "2026-01-01T13:30:00Z"
      }
    ]
  },
  "reviewThreads": {
    "nodes": [
      {
        "isResolved": true,
        "isOutdated": false,
        "comments": {
          "nodes": [
            {
              "databaseId": 300,
              "body": "consider extracting this",
              "author": {"login": "eve"},
              "path": "foo.go",
              "line": 42,
              "url": "https://github.com/org/repo/pull/7#discussion_r300",
              "createdAt": "2026-01-01T14:00:00Z"
            }
          ]
        }
      },
      {
        "isResolved": false,
        "isOutdated": true,
        "comments": {
          "nodes": [
            {
              "databaseId": 301,
              "body": "nit",
              "author": {"login": "frank"},
              "path": "bar.go",
              "line": 7,
              "url": "https://github.com/org/repo/pull/7#discussion_r301",
              "createdAt": "2026-01-01T14:30:00Z"
            }
          ]
        }
      }
    ]
  }
}`

func TestGraphQLPRWithMergedComments(t *testing.T) {
	var gqlPR graphQLPR
	require.NoError(t, json.Unmarshal([]byte(mergedPRFixture), &gqlPR))

	pr := gqlPR.toPRInfo()

	// Issue comments + non-empty review bodies should flow into Comments.
	// Review threads contribute one comment per thread node too.
	require.Len(t, pr.Comments, 4, "1 issue comment + 1 non-empty review + 2 thread comments")

	// Issue comment came through first.
	issueComment := pr.Comments[0]
	assert.Equal(t, int64(100), issueComment.ID)
	assert.Equal(t, "LGTM", issueComment.Body)
	assert.Equal(t, "bob", issueComment.Author)
	assert.Equal(t, time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), issueComment.CreatedAt)

	// The empty-body review (dave) was skipped — matches the old REST behavior.
	// carol's non-empty review body is next.
	reviewBody := pr.Comments[1]
	assert.Equal(t, int64(200), reviewBody.ID)
	assert.Equal(t, "carol", reviewBody.Author)
	assert.Equal(t, "top-level review body", reviewBody.Body)

	// Thread 1 (resolved).
	thread1 := pr.Comments[2]
	assert.Equal(t, int64(300), thread1.ID)
	assert.Equal(t, "foo.go", thread1.Path)
	assert.Equal(t, 42, thread1.Line)
	assert.True(t, thread1.IsResolved)
	assert.False(t, thread1.IsOutdated)

	// Thread 2 (outdated).
	thread2 := pr.Comments[3]
	assert.Equal(t, int64(301), thread2.ID)
	assert.False(t, thread2.IsResolved)
	assert.True(t, thread2.IsOutdated)

	// ReviewThreads should be populated with just the thread-sourced comments.
	require.Len(t, pr.ReviewThreads, 2)
	assert.Equal(t, int64(300), pr.ReviewThreads[0].ID)
	assert.True(t, pr.ReviewThreads[0].IsResolved)
	assert.Equal(t, int64(301), pr.ReviewThreads[1].ID)
	assert.True(t, pr.ReviewThreads[1].IsOutdated)
}

func TestGraphQLPREmptyCommentSections(t *testing.T) {
	// A PR with no comments, reviews, or threads should still unmarshal cleanly
	// and produce an empty Comments/ReviewThreads.
	minimal := `{
	  "number": 1,
	  "title": "x",
	  "author": {"login": "a"},
	  "commits": {"nodes": []},
	  "comments": {"nodes": []},
	  "reviews": {"nodes": []},
	  "reviewThreads": {"nodes": []}
	}`

	var gqlPR graphQLPR
	require.NoError(t, json.Unmarshal([]byte(minimal), &gqlPR))
	pr := gqlPR.toPRInfo()

	assert.Empty(t, pr.Comments)
	assert.Empty(t, pr.ReviewThreads)
}
