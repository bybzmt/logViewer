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
                <li>添加</li>
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
                    <td>换行</td>
                    <td>
                        <select class="border" bind:value={selected.Separator}>
                            <option value={0}>Linux</option>
                            <option value={1}>Windows</option>
                            <option value={2}>mac</option>
                        </select>
                    </td>
                </tr>
                <tr>
                    <td>行首匹配</td>
                    <td><input class="border" bind:value={selected.LineMatch} /></td>
                </tr>
                <tr>
                    <td>过滤器</td>
                    <td><textarea class="border" bind:value={selected.Filter} /></td>
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
