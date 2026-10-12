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

  const host = (state.host || '').trim()
  const port = item.port && item.port > 0 ? item.port : null

  if (!host) {
    // No host configured: fall back to public URL
    return item.url || ''
  }

  try {
    const base = item.url && /^https?:\/\//i.test(item.url) ? item.url : `http://${host}`
    const url = new URL(base)
    url.hostname = host
    if (port)
      url.port = String(port)
    else
      url.port = ''
    return url.toString()
  }
  catch {
    if (port)
      return `http://${host}:${port}`
    return `http://${host}`
  }
}

export function genNetworkStateId(): string {
  return `ns_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`
}
