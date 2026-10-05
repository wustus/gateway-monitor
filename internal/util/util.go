// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package util

import "strings"

func IsWildcardDomain(domain string) bool {
  return strings.HasPrefix(domain, "*.")
}

func ReplaceWildcardDomain(domain, target string) string {
  return strings.Replace(domain, "*", target, 1)
}

// Intersects the hostname of the listener and route, returning the most specific hostname.
//  Returns the most specific hostname and if both hostnames can intersect.
//  See: https://gateway-api.sigs.k8s.io/docs/concepts/hostnames/#hostname-intersection
func GetHostnameIntersection(listenerHostname, routeHostname *string) (string, bool) {
  lh, rh := "*", "*"
  if listenerHostname != nil {
    lh = *listenerHostname
  }
  // route object accepts any hostname
  // listener hostname is at least as specific as that of the route
  if routeHostname == nil {
    return lh, true
  }
  rh = *routeHostname
  // both are equal, they intersect
  if lh == rh {
    return lh, true
  }
  // special case '*', other hostname is at least as specific
  if lh == "*" {
    return rh, true
  }
  if rh == "*" {
    return lh, true
  }
  lw := strings.HasPrefix(lh, "*.")
  rw := strings.HasPrefix(rh, "*.")
  // both hostnames are specific and not equal
  if !lw && !rw {
    return "", false
  }
  // wildcard listener, precise route
  if lw && !rw {
    if strings.HasSuffix(rh, lh[1:]) {
      return rh, true
    }
    return "", false
  }
  // precise listener, wildcard route
  if !lw && rw {
    if strings.HasSuffix(lh, rh[1:]) {
      return lh, true
    }
    return "", false
  }
  // both hostnames are wildcards now we compare the suffix
  ls := lh[1:]
  rs := rh[1:]
  if strings.HasSuffix(ls, rs) {
    return ls, true
  }
  if strings.HasSuffix(rs, ls) {
    return rs, true
  }
  return "", false
}
