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
