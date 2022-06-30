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
        <div class="border w-1/5">
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
            <div>
                <input class="border w-[10em]" placeholder="开始时间" />
                <input class="border w-[10em]" placeholder="结束时间" />
                <input class="border w-[5em]" placeholder="偏移位置" />
                <select>
                    <option>开头</option>
                    <option>未尾</option>
                </select>
                <input class="border w-[5em]" placeholder="显示数量" />
                <button>开始</button>
            </div>
            <div class="border">1</div>
        </div>
    </div>
</Layout>

<style>
</style>
