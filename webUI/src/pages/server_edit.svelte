<script context="module">
    export const load = async ({ query, context }) => {
        let id = query.get("id");

        let row = {};

        if (id) {
            let axios = context.get("axios");

            row = await axios({ url: "/api/server?ID=" + query.get("id") }).then((resp) => {
                return resp.data;
            });
        }

        return {
            props: {
                row: row,
            },
        };
    };
</script>

<script>
    import Layout from "./lib/layout.svelte";
    import { getContext, onMount, onDestroy } from "svelte";

    let axios = getContext("axios");
    let goto = getContext("goto");

    export let row;

    function save() {
        let url = row.ID ? "/api/server/edit" : "/api/server/add";
        axios({
            method: "post",
            url: url,
            data: new URLSearchParams(row),
        }).then(() => {
            cancel();
        });
    }

    function cancel() {
        goto("#/logs?sid=" + row.ServerID);
    }

    onMount(() => {});
</script>

<Layout>
    <div class="border">
        <table>
            <tr>
                <td>名称</td>
                <td><input class="border" bind:value={row.Note} /></td>
            </tr>
            <tr>
                <td>地址</td>
                <td><input class="border" bind:value={row.Addr} /></td>
            </tr>
            <tr>
                <td>用户名</td>
                <td><input class="border" bind:value={row.User} /></td>
            </tr>
            <tr>
                <td>验证方式</td>
                <td>
                    <select class="border" bind:value={row.UsePwd}>
                        <option value={false}>使用密钥</option>
                        <option value={true}>使用密码</option>
                    </select>
                </td>
            </tr>
            {#if row.UsePwd}
                <tr>
                    <td>密码</td>
                    <td><input class="border" bind:value={row.Passwd} /></td>
                </tr>
            {:else}
                <tr>
                    <td>密钥</td>
                    <td><textarea class="border" bind:value={row.PrivateKey} /></td>
                </tr>
            {/if}
            <tr>
                <td colspan="2" class="text-center">
                    <button on:click={save}>保存</button>
                    <button on:click={cancel}>取消</button>
                </td>
            </tr>
        </table>
    </div>
</Layout>

<style>
    input,
    textarea {
        width: 40em;
    }
</style>
