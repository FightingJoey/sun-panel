/** Default network states for a new group. */
export function defaultNetworkConfig(): Panel.NetworkConfig {
  return {
    networkStates: [
      { id: 'wan', name: '公网', type: 'wan', host: '' },
      { id: 'lan', name: '内网', type: 'lan', host: '' },
    ],
    activeNetworkStateId: 'wan',
  }
}

export function ensureNetworkConfig(group?: Panel.ItemIconGroup | null): Panel.NetworkConfig {
  if (group?.networkConfig && group.networkConfig.networkStates?.length)
    return group.networkConfig
  return defaultNetworkConfig()
}

export function getActiveNetworkState(group?: Panel.ItemIconGroup | null): Panel.NetworkState {
  const cfg = ensureNetworkConfig(group)
  return cfg.networkStates.find(s => s.id === cfg.activeNetworkStateId)
    || cfg.networkStates.find(s => s.type === 'wan')
    || cfg.networkStates[0]
}

/**
 * Parse a group network "host" field that users may fill as:
 * - bare IP / hostname: 192.168.1.10, nas.local
 * - host:port: 192.168.1.10:8080
 * - full URL / base: http://192.168.1.10:8080, https://nas.local/path
 * - IPv6: [fd00::1] or [fd00::1]:8080
 *
 * IMPORTANT: never assign raw values with ":" or "://" to URL.hostname —
 * browsers/Node silently ignore invalid hostname sets, leaving the public URL unchanged.
 */
export function parseHostInput(raw: string): { hostname: string; port: string | null; protocol: string | null; pathname: string | null } {
  const trimmed = (raw || '').trim()
  if (!trimmed) {
    return { hostname: '', port: null, protocol: null, pathname: null }
  }

  // Full URL with scheme
  if (/^[a-z][a-z0-9+.-]*:\/\//i.test(trimmed)) {
    try {
      const u = new URL(trimmed)
      return {
        hostname: u.hostname,
        port: u.port || null,
        protocol: u.protocol || null,
        pathname: u.pathname && u.pathname !== '/' ? u.pathname : null,
      }
    }
    catch {
      // fall through to looser parsing
    }
  }

  // [IPv6] or [IPv6]:port
  const ipv6 = trimmed.match(/^\[([^\]]+)](?::(\d+))?$/)
  if (ipv6) {
    return { hostname: ipv6[1], port: ipv6[2] || null, protocol: null, pathname: null }
  }

  // host:port (avoid bare IPv6 which contains multiple colons)
  if (trimmed.includes(':') && !trimmed.includes('::')) {
    const lastColon = trimmed.lastIndexOf(':')
    const maybePort = trimmed.slice(lastColon + 1)
    if (/^\d+$/.test(maybePort)) {
      return {
        hostname: trimmed.slice(0, lastColon),
        port: maybePort,
        protocol: null,
        pathname: null,
      }
    }
  }

  return { hostname: trimmed, port: null, protocol: null, pathname: null }
}

/**
 * Resolve the jump URL for an item under a group's active network state.
 * - wan: use item.url (full public URL)
 * - otherwise: take protocol/path from item.url, replace hostname with state.host and port with item.port
 */
export function buildJumpUrl(item: Panel.ItemInfo, group?: Panel.ItemIconGroup | null): string {
  const state = getActiveNetworkState(group)
  if (!state || state.type === 'wan')
    return item.url || ''

  return buildUrlForState(item, state)
}

/** Build URL for a specific network state (right-click menus). */
export function buildUrlForState(item: Panel.ItemInfo, state: Panel.NetworkState): string {
  if (!state || state.type === 'wan')
    return item.url || ''

  const parsed = parseHostInput(state.host || '')
  if (!parsed.hostname) {
    // No host configured: prefer legacy lanUrl, else public URL
    if (item.lanUrl)
      return item.lanUrl
    return item.url || ''
  }

  const itemPort = item.port && Number(item.port) > 0 ? String(item.port) : null
  // Priority: item.port > host-field port > keep public URL port
  const preferredPort = itemPort || parsed.port

  try {
    const hasPublicUrl = !!(item.url && /^https?:\/\//i.test(item.url))
    const base = hasPublicUrl
      ? item.url!
      : `${parsed.protocol || 'http:'}//${parsed.hostname}`
    const url = new URL(base)

    // hostname only — never pass host:port here (silent no-op otherwise)
    url.hostname = parsed.hostname

    if (preferredPort)
      url.port = preferredPort
    else if (!hasPublicUrl)
      url.port = ''
    // else: keep port from public URL when item/host did not specify one

    if (parsed.protocol)
      url.protocol = parsed.protocol

    // If public URL was empty and host input carried a path, keep it
    if (!hasPublicUrl && parsed.pathname)
      url.pathname = parsed.pathname

    return url.toString()
  }
  catch {
    if (preferredPort)
      return `http://${parsed.hostname}:${preferredPort}`
    return `http://${parsed.hostname}`
  }
}

export function genNetworkStateId(): string {
  return `ns_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`
}
