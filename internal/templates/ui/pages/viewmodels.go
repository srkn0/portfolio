package pages

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/srkn0/main/internal/content"
	i18npkg "github.com/srkn0/main/pkg/i18n"
)

type HomeViewModel struct {
	LatestPosts      []content.PostSummary
	FeaturedProjects []content.ProjectSummary
	TopTags          []content.TagCount
}

func TagsText(tags []string) string {
	return strings.Join(tags, " ")
}

func YearCount(ctx context.Context, count int) string {
	return i18npkg.Tc(ctx, "writing_year_count", map[string]any{"Count": count})
}

func TagCountLabel(tag content.TagCount) string {
	return fmt.Sprintf("%s %d", tag.Name, tag.Count)
}

func TagFilterURL(tag string) string {
	values := url.Values{}
	values.Set("tag", tag)
	return "/blog?" + values.Encode()
}
