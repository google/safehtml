// Copyright (c) 2017 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package safehtml

import "testing"

func TestURLSanitized(t *testing.T) {
	tests := []struct {
		desc    string
		url     string
		wantOut string
	}{
		// Allowed schemes
		{
			desc:    "https public URL",
			url:     "https://example.com/path",
			wantOut: "https://example.com/path",
		},
		{
			desc:    "http public URL",
			url:     "http://example.com/path",
			wantOut: "http://example.com/path",
		},
		{
			desc:    "mailto URL",
			url:     "mailto:user@example.com",
			wantOut: "mailto:user@example.com",
		},
		{
			desc:    "ftp URL",
			url:     "ftp://example.com/file",
			wantOut: "ftp://example.com/file",
		},
		// Relative URLs (no scheme)
		{
			desc:    "absolute path-relative URL",
			url:     "/path/to/resource",
			wantOut: "/path/to/resource",
		},
		{
			desc:    "relative URL",
			url:     "path/to/resource",
			wantOut: "path/to/resource",
		},
		{
			desc:    "fragment-only URL",
			url:     "#section",
			wantOut: "#section",
		},
		// Safe data: URLs with allowed MIME types
		{
			desc:    "data URL with image/png",
			url:     "data:image/png;base64,abc=",
			wantOut: "data:image/png;base64,abc=",
		},
		{
			desc:    "data URL with audio/mpeg",
			url:     "data:audio/mpeg;base64,abc=",
			wantOut: "data:audio/mpeg;base64,abc=",
		},
		// Dangerous schemes that must be blocked to prevent XSS
		{
			desc:    "javascript: scheme blocked",
			url:     "javascript:alert(1)",
			wantOut: InnocuousURL,
		},
		{
			desc:    "JAVASCRIPT: (uppercase) scheme blocked",
			url:     "JAVASCRIPT:alert(1)",
			wantOut: InnocuousURL,
		},
		{
			desc:    "data:text/html scheme causes XSS and must be blocked",
			url:     "data:text/html,<script>alert(1)</script>",
			wantOut: InnocuousURL,
		},
		{
			desc:    "data:text/html;base64 causes XSS and must be blocked",
			url:     "data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==",
			wantOut: InnocuousURL,
		},
		{
			desc:    "vbscript: scheme causes XSS (IE) and must be blocked",
			url:     "vbscript:alert(1)",
			wantOut: InnocuousURL,
		},
		{
			desc:    "blob: scheme must be blocked (can host arbitrary content)",
			url:     "blob:https://example.com/abc",
			wantOut: InnocuousURL,
		},
		{
			desc:    "data:application/javascript causes XSS and must be blocked",
			url:     "data:application/javascript,alert(1)",
			wantOut: InnocuousURL,
		},
		{
			desc:    "data:text/plain must be blocked (non-allowlisted MIME type)",
			url:     "data:text/plain;base64,aGVsbG8=",
			wantOut: InnocuousURL,
		},
		{
			desc:    "unknown scheme must be blocked",
			url:     "unknownscheme:something",
			wantOut: InnocuousURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := URLSanitized(tt.url).String()
			if got != tt.wantOut {
				t.Errorf("URLSanitized(%q) = %q, want %q", tt.url, got, tt.wantOut)
			}
		})
	}
}
