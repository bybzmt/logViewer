<script>
    import Layout from "./lib/layout.svelte";
    import { getContext, onMount, onDestroy } from "svelte";

    let axios = getContext("axios");

    let rows = [];
    let selected = {};

    function viewLogs() {
        axios({ url: "/api/servers" }).then((resp) => {
            rows = resp.data;
        });
    }

    function save() {
        let url = selected.ID ? "/api/server/edit" : "/api/server/add";
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

    onMount(() => {
        viewLogs();
    });
</script>

<Layout>
    <div class="flex">
        <div class="border w-1/4">
            <ul>
                {#each rows as log}
                    <li
                        on:click={() => {
                            selected = log;
                        }}>
                        {log.Note ? log.Note : log.ID}
                    </li>
                {/each}
                <li on:click={add}>添加</li>
            </ul>
        </div>
        <div class="border w-full">
            <table>
                <tr>
                    <td>名称</td>
                    <td><input class="border" bind:value={selected.Note} /></td>
                </tr>
                <tr>
                    <td>地址</td>
                    <td><input class="border" bind:value={selected.Addr} /></td>
                </tr>
                <tr>
                    <td>用户名</td>
                    <td><input class="border" bind:value={selected.User} /></td>
                </tr>
                <tr>
                    <td>验证方式</td>
                    <td>
                        <select class="border" bind:value={selected.UsePwd}>
                            <option value={false}>使用密钥</option>
                            <option value={true}>使用密码</option>
                        </select>
                    </td>
                </tr>
                {#if selected.UsePwd}
                    <tr>
                        <td>密码</td>
                        <td><input class="border" bind:value={selected.Passwd} /></td>
                    </tr>
                {:else}
                    <tr>
                        <td>密钥</td>
                        <td><textarea class="border" bind:value={selected.PrivateKey} /></td>
                    </tr>
                {/if}
                <tr>
                    <td colspan="2" class="text-center">
                        {#if selected.ID}
                            <button on:click={save}>保存</button>
                        {:else}
                            <button on:click={save}>添加</button>
                        {/if}
                    </td>
                </tr>
            </table>
        </div>
    </div>
</Layout>

<style>
</style>
