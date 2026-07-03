package gql

import (
	"testing"

	"github.com/stn1slv/staraudit/pkg/context"
	"github.com/stretchr/testify/assert"
)

func TestCacheEntryFilename(t *testing.T) {
	ctx := &context.Context{
		RepoOwner:          "ullaakut",
		RepoName:           "staraudit",
		GithubToken:        "fakeToken",
		CacheDirectoryPath: "./data",
	}

	sanitizedFilename := cacheEntryFilename(ctx, "https://fakeapi.com/graphql?access_token=fakeToken-1-2019")

	assert.Equal(t, "data/ullaakut/staraudit/https-fakeapi-com-graphql-1-2019", sanitizedFilename)
}
