// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)


func TestIsWildcardDomain(t *testing.T) {
  tests := []struct{
    domain      string
    isWildcard  bool
  }{
    { domain: "example.com", isWildcard: false },
    { domain: "app.example.com", isWildcard: false },
    { domain: "*.example.com", isWildcard: true },
    { domain: "ap*.example.com", isWildcard: false },
    { domain: "*.com", isWildcard: true },
    { domain: "*pp.com", isWildcard: false },
    // it is but we don't want to support '*' at this point
    { domain: "*", isWildcard: false },
  }
  for _, test := range tests {
    res := IsWildcardDomain(test.domain)
    if res != test.isWildcard {
      t.Errorf("IsWildcardDomain('%s') is %v, want %v", test.domain, res, test.isWildcard)
    }
  }
}

func TestGetHostnameIntersection(t *testing.T) {
  tests := []struct {
    lh            string
    rh            string
    intersection  string
    doesIntersect bool
  }{
    // Exact hostname matches
    {lh: "example.com", rh: "example.com", intersection: "example.com", doesIntersect: true},
    {lh: "foo.example.com", rh: "foo.example.com", intersection: "foo.example.com", doesIntersect: true},
    {lh: "foo.bar.example.com", rh: "foo.bar.example.com", intersection: "foo.bar.example.com", doesIntersect: true},

    // Exact hostname mismatches
    {lh: "foo.example.com", rh: "bar.example.com", intersection: "", doesIntersect: false},
    {lh: "example.com", rh: "foo.example.com", intersection: "", doesIntersect: false},
    {lh: "foo.example.com", rh: "example.com", intersection: "", doesIntersect: false},
    {lh: "foo.example.com", rh: "foo.example.net", intersection: "", doesIntersect: false},

    // Wildcard listener, exact route
    {lh: "*.example.com", rh: "foo.example.com", intersection: "foo.example.com", doesIntersect: true},
    {lh: "*.example.com", rh: "bar.example.com", intersection: "bar.example.com", doesIntersect: true},
    {lh: "*.example.com", rh: "foo.bar.example.com", intersection: "foo.bar.example.com", doesIntersect: true},
    {lh: "*.example.com", rh: "example.com", intersection: "", doesIntersect: false},
    {lh: "*.example.com", rh: "foo.example.net", intersection: "", doesIntersect: false},
    {lh: "*.foo.example.com", rh: "bar.foo.example.com", intersection: "bar.foo.example.com", doesIntersect: true},
    {lh: "*.foo.example.com", rh: "foo.example.com", intersection: "", doesIntersect: false},

    // Exact listener, wildcard route
    {lh: "foo.example.com", rh: "*.example.com", intersection: "foo.example.com", doesIntersect: true},
    {lh: "foo.bar.example.com", rh: "*.example.com", intersection: "foo.bar.example.com", doesIntersect: true},
    {lh: "example.com", rh: "*.example.com", intersection: "", doesIntersect: false},
    {lh: "foo.example.net", rh: "*.example.com", intersection: "", doesIntersect: false},
    {lh: "bar.foo.example.com", rh: "*.foo.example.com", intersection: "bar.foo.example.com", doesIntersect: true},
    {lh: "foo.example.com", rh: "*.foo.example.com", intersection: "", doesIntersect: false},

    // Wildcard against wildcard
    {lh: "*.example.com", rh: "*.example.com", intersection: "*.example.com", doesIntersect: true},
    {lh: "*.example.com", rh: "*.foo.example.com", intersection: "*.foo.example.com", doesIntersect: true},
    {lh: "*.foo.example.com", rh: "*.example.com", intersection: "*.foo.example.com", doesIntersect: true},
    {lh: "*.com", rh: "*.example.com", intersection: "*.example.com", doesIntersect: true},
    {lh: "*.example.com", rh: "*.com", intersection: "*.example.com", doesIntersect: true},
    {lh: "*.com", rh: "*.foo.example.com", intersection: "*.foo.example.com", doesIntersect: true},
    {lh: "*.foo.example.com", rh: "*.bar.example.com", intersection: "", doesIntersect: false},
    {lh: "*.example.com", rh: "*.example.net", intersection: "", doesIntersect: false},
    {lh: "*.foo.com", rh: "*.bar.com", intersection: "", doesIntersect: false},

    // Global wildcard (unspecified hostname)
    {lh: "*", rh: "*", intersection: "*", doesIntersect: true},
    {lh: "*", rh: "example.com", intersection: "example.com", doesIntersect: true},
    {lh: "example.com", rh: "*", intersection: "example.com", doesIntersect: true},
    {lh: "*", rh: "*.example.com", intersection: "*.example.com", doesIntersect: true},
    {lh: "*.example.com", rh: "*", intersection: "*.example.com", doesIntersect: true},
    {lh: "*", rh: "foo.bar.example.com", intersection: "foo.bar.example.com", doesIntersect: true},
    {lh: "foo.bar.example.com", rh: "*", intersection: "foo.bar.example.com", doesIntersect: true},

    // Different wildcard depths
    {lh: "*.com", rh: "foo.example.com", intersection: "foo.example.com", doesIntersect: true},
    {lh: "*.com", rh: "example.com", intersection: "example.com", doesIntersect: true},
    {lh: "*.com", rh: "com", intersection: "", doesIntersect: false},
    {lh: "*.example.com", rh: "*.bar.foo.example.com", intersection: "*.bar.foo.example.com", doesIntersect: true},
    {lh: "*.bar.foo.example.com", rh: "*.example.com", intersection: "*.bar.foo.example.com", doesIntersect: true},

    // Boundary conditions
    {lh: "*.example.com", rh: "notexample.com", intersection: "", doesIntersect: false},
    {lh: "*.example.com", rh: "foo.notexample.com", intersection: "", doesIntersect: false},
    {lh: "*.example.com", rh: "foo.example.com.evil", intersection: "", doesIntersect: false},
    {lh: "*.example.com", rh: "fooexample.com", intersection: "", doesIntersect: false},
    {lh: "*.ample.com", rh: "foo.example.com", intersection: "", doesIntersect: false},
  }
  for _, test := range tests {
    intersection, doesIntersect := GetHostnameIntersection(&test.lh, &test.rh)
    if intersection != test.intersection {
      t.Errorf("GetHostnameIntersection('%s', '%s') intersection is '%s', want '%s'", test.lh, test.rh, intersection, test.intersection)
    }
    if doesIntersect != test.doesIntersect {
      t.Errorf("GetHostnameIntersection('%s', '%s') doesIntersect is '%v', want '%v'", test.lh, test.rh, doesIntersect, test.doesIntersect)
    }
  }
}
