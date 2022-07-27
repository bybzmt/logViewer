
export default {
    map: {
        "/": {
            page: () => import('./pages/index.svelte'),
        },
        "/logs": {
            page: () => import('./pages/logs.svelte'),
        },
        "/log_edit": {
            page: () => import('./pages/log_edit.svelte'),
        },
        "/server_edit": {
            page: () => import('./pages/server_edit.svelte'),
        },
        "/logshow": {
            page: () => import('./pages/logshow.svelte'),
        },
    }
}
