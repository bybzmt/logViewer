<script>
    import Layout from "./lib/layout.svelte";
    import { getContext, onMount, onDestroy } from "svelte";

    let axios = getContext("axios");

    let servers = [];

    function getServers() {
        axios({ url: "/api/servers" }).then((resp) => {
            servers = resp.data;
        });
    }

    onMount(() => {
        getServers();
    });
</script>

<Layout>
    <div class="w-1/4">
        <select rows={servers}>
            {#each servers as row}
                <option value="{row}">{row.Note}</option>
            {/each}
        </select>
        <button>logs</button>
    </div>
</Layout>

<style>
</style>
