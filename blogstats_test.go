package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.goblog.app/app/pkgs/contenttype"
)

func Test_blogStats(t *testing.T) {
	app := &goBlog{
		cfg: &config{
			Db: &configDb{
				File: filepath.Join(t.TempDir(), "test.db"),
			},
			Server: &configServer{
				PublicAddress: "https://example.com",
			},
			Blogs: map[string]*configBlog{
				"en": {
					Lang: "en",
					BlogStats: &configBlogStats{
						Enabled: true,
						Path:    "/stats",
					},
					Sections: map[string]*configSection{
						"test": {},
					},
				},
			},
			DefaultBlog: "en",
			User:        &configUser{},
			Webmention: &configWebmention{
				DisableSending: true,
			},
		},
	}

	_ = app.initDatabase(false)
	defer app.db.close()
	app.initComponents(false)

	// Insert post

	err := app.createPost(&post{
		Content:   "This is a simple **test** post",
		Blog:      "en",
		Section:   "test",
		Published: "2020-06-01",
		Status:    statusPublished,
	})
	require.NoError(t, err)

	err = app.createPost(&post{
		Content:   "This is another simple **test** post",
		Blog:      "en",
		Section:   "test",
		Published: "2021-05-01",
		Status:    statusPublished,
	})
	require.NoError(t, err)

	// Test stats

	sd, err := app.db.getBlogStats("en")
	require.NoError(t, err)
	require.NotNil(t, sd)

	require.NotNil(t, sd.Total)
	assert.Equal(t, "2", sd.Total.Posts)
	assert.Equal(t, "12", sd.Total.Words)
	assert.Equal(t, "48", sd.Total.Chars)

	// 2021
	require.NotNil(t, sd.Years)
	row := sd.Years[0]
	require.NotNil(t, row)
	assert.Equal(t, "2021", row.Name)
	assert.Equal(t, "1", row.Posts)
	assert.Equal(t, "6", row.Words)
	assert.Equal(t, "27", row.Chars)

	// 2021-05
	require.NotNil(t, sd.Months)
	require.NotEmpty(t, sd.Months["2021"])
	row = sd.Months["2021"][0]
	require.NotNil(t, row)
	assert.Equal(t, "05", row.Name)
	assert.Equal(t, "1", row.Posts)
	assert.Equal(t, "6", row.Words)
	assert.Equal(t, "27", row.Chars)

	// 2020
	require.NotNil(t, sd.Years)
	row = sd.Years[1]
	require.NotNil(t, row)
	assert.Equal(t, "2020", row.Name)
	assert.Equal(t, "1", row.Posts)
	assert.Equal(t, "6", row.Words)
	assert.Equal(t, "21", row.Chars)

	// Test if cache exists

	assert.NotNil(t, app.db.loadBlogStatsCache("en"))

	// Test HTML

	t.Run("Test stats table", func(t *testing.T) {
		h := http.HandlerFunc(app.serveBlogStatsTable)

		req := httptest.NewRequest(http.MethodGet, "/abc", nil)
		req = req.WithContext(context.WithValue(req.Context(), blogKey, "en"))

		rec := httptest.NewRecorder()

		h(rec, req)

		res := rec.Result()
		resBody, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		resString := string(resBody)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, resString, "class=statsyear data-year=2021")
		assert.Contains(t, res.Header.Get(contentType), contenttype.HTML)
	})

	t.Run("Test stats page", func(t *testing.T) {
		h := http.HandlerFunc(app.serveBlogStats)

		req := httptest.NewRequest(http.MethodGet, "/abc", nil)
		req = req.WithContext(context.WithValue(req.Context(), blogKey, "en"))

		rec := httptest.NewRecorder()

		h(rec, req)

		res := rec.Result()
		resBody, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		resString := string(resBody)

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, resString, "data-table=/stats.table.html")
		assert.Contains(t, res.Header.Get(contentType), contenttype.HTML)
	})

}
