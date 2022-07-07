<script>
    import Layout from "./lib/layout.svelte";
    import { getContext, onMount, onDestroy } from "svelte";

    let axios = getContext("axios");
    let rows = [];
    let selected = {};

    let ws;
    let msgs = [];

    function viewLogs() {
        axios({ url: "/api/viewLogs" }).then((resp) => {
            rows = resp.data;
        });
    }

    function search() {
        msgs = [];

        ws = new WebSocket("ws://" + API_BASE + "/api/logs");
        ws.onopen = () => {
            console.log("open");

            for (let i = 0; i < 1000; i++) {
                ws.send(i);
            }
        };
        ws.onclose = () => {
            console.log("close");
            ws = null;
        };
        ws.onerror = () => {
            console.log("error");
        };
        ws.onmessage = (evt) => {
            console.log("onmessage");

            msgs.push(evt.data);
            if (msgs.length > 1000) {
                msgs = msgs.slice(-1000);
            } else {
                msgs = msgs;
            }
        };
    }
    function cancel() {
        ws.close();
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
                <select>
                    <option>开头</option>
                    <option>未尾</option>
                </select>
                <input class="border w-[5em]" placeholder="显示数量" />
                <input class="border w-[5em]" placeholder="匹配" />
                {#if ws == null}
                    <button on:click={search}>开始</button>
                {:else}
                    <button on:click={cancel}>取消</button>
                {/if}
            </div>
            <div class="border">
                {#each msgs as msg}
                    <p>{msg}</p>
                {/each}
            </div>
        </div>
    </div>
</Layout>

<style>
</style>
