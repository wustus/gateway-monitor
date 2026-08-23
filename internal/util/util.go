package util

import "strings"

func IsWildcardDomain(domain string) bool {
  return strings.HasPrefix(domain, "*.")
}

func ReplaceWildcardDomain(domain, target string) string {
  return strings.Replace(domain, "*", target, 1)
}
