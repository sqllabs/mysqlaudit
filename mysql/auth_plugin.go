package mysql

import "strings"

// SupportedAuthPlugins lists plugins goInception can negotiate with clients.
var SupportedAuthPlugins = map[string]struct{}{
	AuthNativePassword:      {},
	AuthCachingSha2Password: {},
}

// NormalizeAuthPlugin normalizes plugin name to lowercase and validates it.
// Unsupported names fall back to AuthDefaultPlugin.
func NormalizeAuthPlugin(name string) string {
	plugin := strings.ToLower(strings.TrimSpace(name))
	if plugin == "" {
		return AuthDefaultPlugin
	}
	if _, ok := SupportedAuthPlugins[plugin]; ok {
		return plugin
	}
	return AuthDefaultPlugin
}
