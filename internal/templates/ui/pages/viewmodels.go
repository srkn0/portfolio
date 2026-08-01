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
	return i18npkg.Tc(ctx, "posts_year_count", map[string]any{"Count": count})
}

func TagCountLabel(tag content.TagCount) string {
	return fmt.Sprintf("%s %d", tag.Name, tag.Count)
}

func TagFilterURL(tag string) string {
	values := url.Values{}
	values.Set("tag", tag)
	return "/blog?" + values.Encode()
}

func ProjectCategoryLabel(ctx context.Context, category string) string {
	if category == "" {
		category = "lab"
	}
	key := "project_category_" + strings.ReplaceAll(category, "-", "_")
	return i18npkg.T(ctx, key)
}
