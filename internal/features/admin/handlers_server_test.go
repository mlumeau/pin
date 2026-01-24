package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseUserListParamsDefaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/settings/admin/server", nil)
	query, sort, dir, page := parseUserListParams(req)
	if query != "" {
		t.Fatalf("expected empty query, got %q", query)
	}
	if sort != "id" {
		t.Fatalf("expected sort id, got %q", sort)
	}
	if dir != "desc" {
		t.Fatalf("expected dir desc, got %q", dir)
	}
	if page != 1 {
		t.Fatalf("expected page 1, got %d", page)
	}
}

func TestParseUserListParamsSortDefaultsDir(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/settings/admin/server?sort=handle", nil)
	_, sort, dir, _ := parseUserListParams(req)
	if sort != "handle" {
		t.Fatalf("expected sort handle, got %q", sort)
	}
	if dir != "asc" {
		t.Fatalf("expected dir asc for sorted list, got %q", dir)
	}
}

func TestParseUserListParamsDirOverride(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/settings/admin/server?sort=handle&dir=desc", nil)
	_, sort, dir, _ := parseUserListParams(req)
	if sort != "handle" {
		t.Fatalf("expected sort handle, got %q", sort)
	}
	if dir != "desc" {
		t.Fatalf("expected dir desc override, got %q", dir)
	}
}

func TestPageBoundsEmptyTotal(t *testing.T) {
	prev, next, total := pageBounds(3, 10, 0)
	if prev != 1 || next != 1 || total != 1 {
		t.Fatalf("expected 1/1/1, got %d/%d/%d", prev, next, total)
	}
}

func TestPageSliceClamps(t *testing.T) {
	start, end := pageSlice(3, 10, 15)
	if start != 15 || end != 15 {
		t.Fatalf("expected clamped 15..15, got %d..%d", start, end)
	}
}
