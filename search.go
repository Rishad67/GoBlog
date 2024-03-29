package main

import (
	"context"
	"encoding/base64"
	"net/http"
	"path"
)

const defaultSearchPath = "/search"
const searchPlaceholder = "{search}"

func (a *goBlog) serveSearch(w http.ResponseWriter, r *http.Request) {
	servePath := r.Context().Value(pathKey).(string)
	err := r.ParseForm()
	if err != nil {
		a.serveError(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	if q := r.Form.Get("q"); q != "" {
		// Clean query
		q = cleanHTMLText(q)
		// Redirect to results
		http.Redirect(w, r, path.Join(servePath, searchEncode(q)), http.StatusFound)
		return
	}
	a.render(w, r, a.renderSearch, &renderData{
		Canonical: a.getFullAddress(servePath),
	})
}

func (a *goBlog) serveSearchResult(w http.ResponseWriter, r *http.Request) {
	a.serveIndex(w, r.WithContext(context.WithValue(r.Context(), indexConfigKey, &indexConfig{
		path: r.Context().Value(pathKey).(string) + "/" + searchPlaceholder,
	})))
}

func searchEncode(search string) string {
	return base64.URLEncoding.EncodeToString([]byte(search))
}

func searchDecode(encoded string) string {
	db, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}
	return string(db)
}
