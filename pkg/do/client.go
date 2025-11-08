package do

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"digitalocean-registry-cleaner/pkg/detect"
)

type DigitalOceanClient struct {
	token     string
	protected []string
	client    *http.Client
}

type Tag struct {
	Tag            string    `json:"tag"`
	ManifestDigest string    `json:"manifest_digest"`
	CompressedSize int       `json:"compressed_size_bytes"`
	Size           int       `json:"size_bytes"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CleanupInput struct {
	Registry   string
	Repository string
	DryRun     bool
	KeepTags   int
	MinAge     time.Duration
}

func NewClient(token string, protected []string) *DigitalOceanClient {
	return &DigitalOceanClient{
		token:     token,
		protected: protected,
		client:    http.DefaultClient,
	}
}

// RunCleanup deletes outdated tags and branches from the registry.
// Returns a list of deleted tags.
func (c *DigitalOceanClient) RunCleanup(input CleanupInput) ([]Tag, error) {
	tags, err := c.listTags(input.Registry, input.Repository)
	if err != nil {
		return nil, fmt.Errorf("could not list tags: %w", err)
	}

	// categorize tags - exceptions, tags, branches
	var keepTags []Tag
	var deleteTags []Tag
	for _, tag := range tags {
		if c.isProtected(tag.Tag) {
			continue // exceptions - never delete
		} else if detect.IsTag(tag.Tag) {
			keepTags = append(keepTags, tag) // git tags
		} else if tag.UpdatedAt.After(time.Now().Add(-input.MinAge)) {
			continue // tag is newer than the minimum age
		} else {
			deleteTags = append(deleteTags, tag) // git branches
		}
	}

	// Sort tags by date
	slices.SortFunc(keepTags, func(a, b Tag) int {
		return a.UpdatedAt.Compare(b.UpdatedAt)
	})

	// Keep the latest N tags
	var releaseTagsToDelete []Tag

	if input.KeepTags > 0 && len(keepTags) > input.KeepTags {
		releaseTagsToDelete = keepTags[0 : len(keepTags)-input.KeepTags]
	}

	var deletedTags []Tag

	// Delete outdated tags
	for _, tag := range releaseTagsToDelete {
		if !input.DryRun {
			if err := c.deleteTag(input.Registry, input.Repository, tag.Tag); err != nil {
				return deletedTags, fmt.Errorf("could not delete tag %s.%s:%s : %w", input.Registry, input.Repository, tag.Tag, err)
			}
		}
		deletedTags = append(deletedTags, tag)
	}

	// Delete outdated branches
	for _, tag := range deleteTags {
		if !input.DryRun {
			if err := c.deleteTag(input.Registry, input.Repository, tag.Tag); err != nil {
				return deletedTags, fmt.Errorf("could not delete tag %s.%s:%s : %w", input.Registry, input.Repository, tag.Tag, err)
			}
		}
		deletedTags = append(deletedTags, tag)
	}

	return deletedTags, nil
}

func (c *DigitalOceanClient) isProtected(tag string) bool {
	for _, protectedTag := range c.protected {
		if strings.EqualFold(protectedTag, tag) {
			return true
		}
	}
	return false
}

func (c *DigitalOceanClient) listTags(registry, repository string) ([]Tag, error) {
	const addr = "https://api.digitalocean.com/v2/registry/%s/repositories/%s/tags"

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(addr, url.PathEscape(registry), url.PathEscape(repository)),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("User-Agent", "digitalocean-registry-cleaner")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	var output = struct {
		Tags []Tag `json:"tags"`
	}{}
	err = json.Unmarshal(body, &output)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %w", err)
	}

	return output.Tags, nil
}

func (c *DigitalOceanClient) deleteTag(registry, repository, tag string) error {
	const addr = "https://api.digitalocean.com/v2/registry/%s/repositories/%s/tags/%s"

	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf(addr, url.PathEscape(registry), url.PathEscape(repository), url.PathEscape(tag)),
		nil,
	)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("User-Agent", "digitalocean-registry-cleaner")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
