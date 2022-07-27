<script context="module">
    export const load = async ({ query, context }) => {
        let axios = context.get("axios");

        let id = query.get("id");
        let server_id = query.get("sid");

        let row = {};

        if (id) {
            let axios = context.get("axios");

            row = await axios({ url: "/api/log?ID=" + query.get("id") }).then((resp) => {
                return resp.data;
            });
        }

        return {
            props: {
                row: row,
                server_id,
            },
        };
    };
</script>

<script>
    import Layout from "./lib/layout.svelte";
    import { getContext, onMount, onDestroy } from "svelte";

    export let row;
    export let server_id;

    let axios = getContext("axios");
    let goto = getContext("goto");

    if (!row.ServerID) {
        row.ServerID = server_id;
    }

    function save() {
        let url = row.ID ? "/api/viewLog/edit" : "/api/viewLog/add";
        axios({
            method: "post",
            url: url,
            data: new URLSearchParams(row),
        }).then(() => {
            goto("#/logs?sid="+server_id);
        });
    }

    function cancel() {
        goto("#/logs?sid="+server_id);
    }

    onMount(() => {});
</script>

<Layout>
    <table>
        <tr>
            <td>名称</td>
            <td><input class="border" bind:value={row.Note} /></td>
        </tr>
        <tr>
            <td>文件列表</td>
            <td><textarea class="border" bind:value={row.Files} /></td>
        </tr>
        <tr>
            <td>时间正则</td>
            <td><input class="border" bind:value={row.TimeRegex} /></td>
        </tr>
        <tr>
            <td>时间格式</td>
            <td><input class="border" bind:value={row.TimeLayout} /></td>
        </tr>
        <tr>
            <td>包含</td>
            <td><textarea class="border" bind:value={row.Contains} /></td>
        </tr>
        <tr>
            <td>正则</td>
            <td><textarea class="border" bind:value={row.Regex} /></td>
        </tr>
        <tr>
            <td colspan="2" class="text-center">
                <button on:click={save}>保存</button>
                <button on:click={cancel}>取消</button>
            </td>
        </tr>
    </table>
</Layout>

<style>
</style>
