<script>
    import Layout from "./lib/layout.svelte";
    import { getContext, onMount, onDestroy } from "svelte";

    let axios = getContext("axios");
    let rows = [];
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
                    <td>文件列表</td>
                    <td><textarea class="border" bind:value={selected.Files} /></td>
                </tr>
                <tr>
                    <td>时间正则</td>
                    <td><input class="border" bind:value={selected.TimeRegex} /></td>
                </tr>
                <tr>
                    <td>时间格式</td>
                    <td><input class="border" bind:value={selected.TimeLayout} /></td>
                </tr>
                <tr>
                    <td>包含</td>
                    <td><textarea class="border" bind:value={selected.Contains} /></td>
                </tr>
                <tr>
                    <td>正则</td>
                    <td><textarea class="border" bind:value={selected.Regex} /></td>
                </tr>
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
