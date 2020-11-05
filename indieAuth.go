package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func checkIndieAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearerToken := r.Header.Get("Authorization")
		if len(bearerToken) == 0 {
			bearerToken = r.URL.Query().Get("access_token")
		}
		tokenData, err := verifyIndieAuthToken(bearerToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if err := compareHostnames(tokenData.Me, appConfig.Server.Domain); err != nil {
			http.Error(w, "Forbidden", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "scope", strings.Join(tokenData.Scopes, " "))
		next.ServeHTTP(w, r.WithContext(ctx))
		return
	})
}

func compareHostnames(a string, allowed string) error {
	h1, err := url.Parse(a)
	if err != nil {
		return err
	}
	if strings.ToLower(h1.Hostname()) != strings.ToLower(allowed) {
		return fmt.Errorf("hostnames do not match, %s is not %s", h1, allowed)
	}
	return nil
}