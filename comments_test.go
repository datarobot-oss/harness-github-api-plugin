package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/v62/github"
)

// newTestClient returns a go-github client that talks to a local fake API served by mux.
func newTestClient(t *testing.T, mux *http.ServeMux) *github.Client {
	t.Helper()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client := github.NewClient(nil)
	baseURL, _ := url.Parse(server.URL + "/")
	client.BaseURL = baseURL
	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatal(err)
	}
}

func readBody(t *testing.T, r *http.Request) string {
	t.Helper()
	var c github.IssueComment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		t.Fatal(err)
	}
	return c.GetBody()
}

func comment(id int64, login string, body string) map[string]any {
	return map[string]any{
		"id":       id,
		"body":     body,
		"html_url": fmt.Sprintf("https://github.com/o/r/pull/7#issuecomment-%d", id),
		"user":     map[string]any{"login": login},
	}
}

func TestUpsertPrCommentWithoutMarkerCreates(t *testing.T) {
	mux := http.NewServeMux()
	var created string
	mux.HandleFunc("POST /repos/o/r/issues/7/comments", func(w http.ResponseWriter, r *http.Request) {
		created = readBody(t, r)
		writeJSON(t, w, comment(11, "bot", created))
	})
	ctx := context.Background()

	fields := upsertPrComment(newTestClient(t, mux), &ctx, "r", "o", "7", "hello", "")

	if created != "hello" {
		t.Errorf("created body = %q, want %q", created, "hello")
	}
	if fields["COMMENT_ID"] != "11" || fields["COMMENT_ACTION"] != "created" {
		t.Errorf("unexpected fields %v", fields)
	}
	if !strings.HasSuffix(fields["COMMENT_URL"], "#issuecomment-11") {
		t.Errorf("unexpected COMMENT_URL %q", fields["COMMENT_URL"])
	}
}

func TestUpsertPrCommentUpdatesOwnMarkedComment(t *testing.T) {
	tag := commentMarkerTag("deploy-eks")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{"login": "bot"})
	})
	mux.HandleFunc("GET /repos/o/r/issues/7/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			writeJSON(t, w, []any{comment(30, "bot", "newer\n\n"+tag)})
			return
		}
		// First page links to a second one to exercise pagination.
		w.Header().Set("Link", fmt.Sprintf(`<http://%s/repos/o/r/issues/7/comments?page=2>; rel="next"`, r.Host))
		writeJSON(t, w, []any{
			comment(10, "bot", "older\n\n"+tag),
			comment(20, "someone", "quoted\n\n"+tag),                              // other author: must be ignored
			comment(25, "bot", "other marker\n\n"+commentMarkerTag("deploy-gke")), // other marker
		})
	})
	var edited string
	mux.HandleFunc("PATCH /repos/o/r/issues/comments/30", func(w http.ResponseWriter, r *http.Request) {
		edited = readBody(t, r)
		writeJSON(t, w, comment(30, "bot", edited))
	})
	mux.HandleFunc("POST /repos/o/r/issues/7/comments", func(w http.ResponseWriter, r *http.Request) {
		t.Error("must not create a new comment when a marked one exists")
	})
	ctx := context.Background()

	fields := upsertPrComment(newTestClient(t, mux), &ctx, "r", "o", "7", "status: done", "deploy-eks")

	if edited != "status: done\n\n"+tag {
		t.Errorf("edited body = %q", edited)
	}
	if fields["COMMENT_ID"] != "30" || fields["COMMENT_ACTION"] != "updated" {
		t.Errorf("unexpected fields %v", fields)
	}
}

func TestUpsertPrCommentCreatesMarkedCommentWhenNoneExists(t *testing.T) {
	tag := commentMarkerTag("deploy-eks")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{"login": "bot"})
	})
	mux.HandleFunc("GET /repos/o/r/issues/7/comments", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, []any{comment(20, "someone", "quoted\n\n"+tag)})
	})
	var created string
	mux.HandleFunc("POST /repos/o/r/issues/7/comments", func(w http.ResponseWriter, r *http.Request) {
		created = readBody(t, r)
		writeJSON(t, w, comment(40, "bot", created))
	})
	ctx := context.Background()

	fields := upsertPrComment(newTestClient(t, mux), &ctx, "r", "o", "7", "deploying", "deploy-eks")

	if created != "deploying\n\n"+tag {
		t.Errorf("created body = %q", created)
	}
	if fields["COMMENT_ID"] != "40" || fields["COMMENT_ACTION"] != "created" {
		t.Errorf("unexpected fields %v", fields)
	}
}

func TestResolveCommentBody(t *testing.T) {
	if got := resolveCommentBody("inline", ""); got != "inline" {
		t.Errorf("got %q, want inline", got)
	}
	file := filepath.Join(t.TempDir(), "body.md")
	if err := os.WriteFile(file, []byte("line 1\nline 2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := resolveCommentBody("inline", file); got != "line 1\nline 2\n" {
		t.Errorf("file body not preferred, got %q", got)
	}
}

func TestAddReaction(t *testing.T) {
	mux := http.NewServeMux()
	var content string
	mux.HandleFunc("POST /repos/o/r/issues/comments/123/reactions", func(w http.ResponseWriter, r *http.Request) {
		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		content = req["content"]
		writeJSON(t, w, map[string]any{"id": 1, "content": content})
	})
	ctx := context.Background()

	addReaction(newTestClient(t, mux), &ctx, "r", "o", "123", "eyes")

	if content != "eyes" {
		t.Errorf("reaction = %q, want eyes", content)
	}
}

func TestGetUserPermission(t *testing.T) {
	cases := map[string]struct {
		status     int
		permission string
		wantPerm   string
		wantWrite  string
	}{
		"admin":   {http.StatusOK, "admin", "admin", "true"},
		"write":   {http.StatusOK, "write", "write", "true"},
		"read":    {http.StatusOK, "read", "read", "false"},
		"none":    {http.StatusOK, "none", "none", "false"},
		"unknown": {http.StatusNotFound, "", "none", "false"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("GET /repos/o/r/collaborators/alice/permission", func(w http.ResponseWriter, r *http.Request) {
				if tc.status != http.StatusOK {
					w.WriteHeader(tc.status)
					writeJSON(t, w, map[string]any{"message": "Not Found"})
					return
				}
				writeJSON(t, w, map[string]any{"permission": tc.permission, "user": map[string]any{"login": "alice"}})
			})
			ctx := context.Background()

			fields := getUserPermission(newTestClient(t, mux), &ctx, "r", "o", "alice")

			if fields["PERMISSION"] != tc.wantPerm || fields["HAS_WRITE_ACCESS"] != tc.wantWrite {
				t.Errorf("got %v, want PERMISSION=%s HAS_WRITE_ACCESS=%s", fields, tc.wantPerm, tc.wantWrite)
			}
		})
	}
}
