
export default {
    map: {
        "/": {
            page: () => import('./pages/index.svelte'),
        },
        "/logs": {
            page: () => import('./pages/logs.svelte'),
        },
        "/servers": {
            page: () => import('./pages/servers.svelte'),
        },
    }
}
