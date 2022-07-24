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
    <table>
        <thead>
            <tr>
                <th>ID</th>
                <th>Note</th>
                <th>Addr</th>
                <th colspan="3"></th>
            </tr>
        </thead>
        <tbody>
            {#each servers as row}
            <tr>
                <td>{row.ID}</td>
                <td>{row.Note}</td>
                <td>{row.Addr}</td>
                <td>
                    <a href="#/logs?sid={row.ID}">Logs</a>
                </td>
                <td>
                    <a href="#/logshow?sid={row.ID}">Edit</a>
                </td>
                <td>
                    <a>Del</a>
                </td>
            </tr>
            {/each}

            <tr>
                <td class="text-center" colspan="5">--</td>
                <td>
                    <a>Add</a>
                </td>
            </tr>
        </tbody>
    </table>
</Layout>

<style>
    table{
        margin: 1em;
    }
    td, th {
        border: 1px solid #777;
    }
</style>
