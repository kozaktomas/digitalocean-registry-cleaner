package do

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockHTTPClient is a mock implementation of http.Client
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

// mockRoundTripper implements http.RoundTripper for mocking
type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestRunCleanup_DeleteOldBranches(t *testing.T) {
	// Create a client with mocked HTTP transport
	client := NewClient("test-token", []string{"prod-protected"})
	now := time.Now()

	// Mock the HTTP client
	client.client = &http.Client{
		Transport: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				// Mock listTags response
				if req.Method == http.MethodGet {
					responseBody := `{
						"tags": [
							{
								"tag": "prod-protected",
								"manifest_digest": "sha256:test",
								"compressed_size_bytes": 12345678,
								"size_bytes": 12345678,
								"updated_at": "` + now.Add(-40*24*time.Hour).Format(time.RFC3339) + `"
							},
							{
								"tag": "tag-40-days-old",
								"manifest_digest": "sha256:test",
								"compressed_size_bytes": 12345678,
								"size_bytes": 12345678,
								"updated_at": "` + now.Add(-40*24*time.Hour).Format(time.RFC3339) + `"
							},
							{
								"tag": "tag-8-days-old",
								"manifest_digest": "sha256:test",
								"compressed_size_bytes": 12345678,
								"size_bytes": 12345678,
								"updated_at": "` + now.Add(-8*24*time.Hour).Format(time.RFC3339) + `"
							},
							{
								"tag": "tag-3-days-old",
								"manifest_digest": "sha256:test",
								"compressed_size_bytes": 12345678,
								"size_bytes": 12345678,
								"updated_at": "` + now.Add(-3*24*time.Hour).Format(time.RFC3339) + `"
							}
						],
						"meta": {
							"total": 3
						}
					}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
						Header:     make(http.Header),
					}, nil
				}

				// Mock delete response
				if req.Method == http.MethodDelete {
					return &http.Response{
						StatusCode: http.StatusNoContent,
						Body:       io.NopCloser(bytes.NewBufferString("")),
						Header:     make(http.Header),
					}, nil
				}

				return nil, nil
			},
		},
	}

	input := CleanupInput{
		Registry:   "test",
		Repository: "test",
		DryRun:     false,
		KeepTags:   0,
		MinAge:     7 * 24 * time.Hour, // 7 day
	}

	deletedTags, err := client.RunCleanup(input)

	assert.NoError(t, err)
	// Should delete 2 tags: "docker" and "tk-docker-versions" (both are branches, older than 1 day)
	// "prod" is protected and should not be deleted
	assert.Equal(t, 2, len(deletedTags))

	// Verify the correct tags were deleted
	tagNames := make([]string, len(deletedTags))
	for i, tag := range deletedTags {
		tagNames[i] = tag.Tag
	}
	assert.NotContains(t, tagNames, "prod-protected")
	assert.NotContains(t, tagNames, "tag-3-days-old")
	assert.Contains(t, tagNames, "tag-40-days-old")
	assert.Contains(t, tagNames, "tag-8-days-old")
}

func TestRunCleanup_KeepLatestNTags(t *testing.T) {
	client := NewClient("test-token", []string{})

	client.client = &http.Client{
		Transport: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodGet {
					responseBody := `{
						"tags": [
							{
								"tag": "1.0.0",
								"manifest_digest": "sha256:abc1",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-11-01T10:00:00Z"
							},
							{
								"tag": "1.1.0",
								"manifest_digest": "sha256:abc2",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-11-02T10:00:00Z"
							},
							{
								"tag": "1.2.0",
								"manifest_digest": "sha256:abc3",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-11-03T10:00:00Z"
							},
							{
								"tag": "1.3.0",
								"manifest_digest": "sha256:abc4",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-11-04T10:00:00Z"
							},
							{
								"tag": "1.4.0",
								"manifest_digest": "sha256:abc5",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-11-05T10:00:00Z"
							}
						]
					}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
						Header:     make(http.Header),
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(bytes.NewBufferString("")),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	input := CleanupInput{
		Registry:   "test",
		Repository: "test",
		DryRun:     false,
		KeepTags:   3, // Keep only the latest 3 tags
		MinAge:     0,
	}

	deletedTags, err := client.RunCleanup(input)

	assert.NoError(t, err)
	// Should delete 2 oldest tags (1.0.0 and 1.1.0), keeping the latest 3 (1.2.0, 1.3.0, 1.4.0)
	assert.Equal(t, 2, len(deletedTags))

	tagNames := make([]string, len(deletedTags))
	for i, tag := range deletedTags {
		tagNames[i] = tag.Tag
	}
	assert.Contains(t, tagNames, "1.0.0")
	assert.Contains(t, tagNames, "1.1.0")
}

func TestRunCleanup_DryRun(t *testing.T) {
	client := NewClient("test-token", []string{})

	deleteCallCount := 0

	client.client = &http.Client{
		Transport: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodGet {
					responseBody := `{
						"tags": [
							{
								"tag": "old-branch",
								"manifest_digest": "sha256:abc123",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-10-01T10:00:00Z"
							}
						]
					}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
						Header:     make(http.Header),
					}, nil
				}

				if req.Method == http.MethodDelete {
					deleteCallCount++
					return &http.Response{
						StatusCode: http.StatusNoContent,
						Body:       io.NopCloser(bytes.NewBufferString("")),
						Header:     make(http.Header),
					}, nil
				}

				return nil, nil
			},
		},
	}

	input := CleanupInput{
		Registry:   "test",
		Repository: "test",
		DryRun:     true, // Dry run mode
		KeepTags:   0,
		MinAge:     24 * time.Hour,
	}

	deletedTags, err := client.RunCleanup(input)

	assert.NoError(t, err)
	// Should identify 1 tag for deletion
	assert.Equal(t, 1, len(deletedTags))
	// But no actual DELETE requests should be made
	assert.Equal(t, 0, deleteCallCount)
}

func TestRunCleanup_ProtectedTags(t *testing.T) {
	client := NewClient("test-token", []string{"main", "prod", "staging"})

	client.client = &http.Client{
		Transport: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodGet {
					responseBody := `{
						"tags": [
							{
								"tag": "main",
								"manifest_digest": "sha256:abc1",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-10-01T10:00:00Z"
							},
							{
								"tag": "prod",
								"manifest_digest": "sha256:abc2",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-10-01T10:00:00Z"
							},
							{
								"tag": "staging",
								"manifest_digest": "sha256:abc3",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-10-01T10:00:00Z"
							},
							{
								"tag": "feature-branch",
								"manifest_digest": "sha256:abc4",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-10-01T10:00:00Z"
							}
						]
					}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
						Header:     make(http.Header),
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(bytes.NewBufferString("")),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	input := CleanupInput{
		Registry:   "test",
		Repository: "test",
		DryRun:     false,
		KeepTags:   0,
		MinAge:     24 * time.Hour,
	}

	deletedTags, err := client.RunCleanup(input)

	assert.NoError(t, err)
	// Should only delete "feature-branch", protected tags should be kept
	assert.Equal(t, 1, len(deletedTags))
	assert.Equal(t, "feature-branch", deletedTags[0].Tag)
}

func TestRunCleanup_MixedTagsAndBranches(t *testing.T) {
	client := NewClient("test-token", []string{})

	client.client = &http.Client{
		Transport: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodGet {
					responseBody := `{
						"tags": [
							{
								"tag": "1.0.0",
								"manifest_digest": "sha256:tag1",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-10-01T10:00:00Z"
							},
							{
								"tag": "1.1.0",
								"manifest_digest": "sha256:tag2",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-10-02T10:00:00Z"
							},
							{
								"tag": "1.2.0",
								"manifest_digest": "sha256:tag3",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-10-03T10:00:00Z"
							},
							{
								"tag": "develop",
								"manifest_digest": "sha256:branch1",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-10-01T10:00:00Z"
							},
							{
								"tag": "feature-x",
								"manifest_digest": "sha256:branch2",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "2025-10-01T10:00:00Z"
							}
						]
					}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
						Header:     make(http.Header),
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(bytes.NewBufferString("")),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	input := CleanupInput{
		Registry:   "test",
		Repository: "test",
		DryRun:     false,
		KeepTags:   2, // Keep only the latest 2 release tags
		MinAge:     24 * time.Hour,
	}

	deletedTags, err := client.RunCleanup(input)

	assert.NoError(t, err)
	// Should delete:
	// - 1 old release tag (1.0.0)
	// - 2 old branches (develop, feature-x)
	// Total: 3 deletions
	assert.Equal(t, 3, len(deletedTags))

	tagNames := make([]string, len(deletedTags))
	for i, tag := range deletedTags {
		tagNames[i] = tag.Tag
	}
	assert.Contains(t, tagNames, "1.0.0")
	assert.Contains(t, tagNames, "develop")
	assert.Contains(t, tagNames, "feature-x")
	assert.NotContains(t, tagNames, "1.1.0")
	assert.NotContains(t, tagNames, "1.2.0")
}

func TestRunCleanup_NoTagsToDelete(t *testing.T) {
	client := NewClient("test-token", []string{})

	now := time.Now()

	client.client = &http.Client{
		Transport: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method == http.MethodGet {
					responseBody := `{
						"tags": [
							{
								"tag": "1.0.0",
								"manifest_digest": "sha256:abc123",
								"compressed_size_bytes": 100000,
								"size_bytes": 200000,
								"updated_at": "` + now.Format(time.RFC3339) + `"
							}
						]
					}`
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
						Header:     make(http.Header),
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(bytes.NewBufferString("")),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	input := CleanupInput{
		Registry:   "test",
		Repository: "test",
		DryRun:     false,
		KeepTags:   1,
		MinAge:     24 * time.Hour,
	}

	deletedTags, err := client.RunCleanup(input)

	assert.NoError(t, err)
	// All tags are recent, nothing should be deleted
	assert.Equal(t, 0, len(deletedTags))
}
