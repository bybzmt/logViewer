<script context="module">
    export const load = async ({ query, context }) => {
        let axios = context.get("axios");

        let rows = await axios({ url: "/api/viewLogs?sid=" + query.get("sid") }).then((resp) => {
            return resp.data;
        });

        return {
            props: {
                rows: rows,
                server_id: query.get("sid"),
            },
        };
    };
</script>

<script>
    import Layout from "./lib/layout.svelte";
    import { getContext, onMount, onDestroy } from "svelte";

    export let rows;
    export let server_id;

    let axios = getContext("axios");
    let selected = {};

    function viewLogs() {
        axios({ url: "/api/viewLogs" }).then((resp) => {
            rows = resp.data;
        });
    }

    function save() {
        let url = selected.ID ? "/api/viewLog/edit" : "/api/viewLog/add";
        axios({
            method: "post",
            url: url,
            data: new URLSearchParams(selected),
        }).then(() => {
            viewLogs();
            selected = {};
        });
    }

    function add() {
        selected = {};
    }

    onMount(() => {});
</script>

<Layout>
    <table>
        <thead>
            <tr>
                <th>ID</th>
                <th>Note</th>
                <th>Files</th>
                <th colspan="3" />
            </tr>
        </thead>
        <tbody>
            {#each rows as row}
                <tr>
                    <td>{row.ID}</td>
                    <td>{row.Note}</td>
                    <td>{row.Files}</td>
                    <td>
                        <a href="#/server_edit?id={row.ID}">Edit</a>
                    </td>
                    <td>
                        <a>Del</a>
                    </td>
                </tr>
            {/each}

            <tr>
                <td class="text-center" colspan="3">--</td>
                <td>
                    <a href="#/log_edit?sid={server_id}">Add</a>
                </td>
                <td>
                    <a href="#/logshow?sid={server_id}">Cancel</a>
                </td>
            </tr>
        </tbody>
    </table>
</Layout>

<style>
    table {
        margin: 1em;
    }
    td,
    th {
        border: 1px solid #777;
        padding: 3px 5px;
    }
</style>
