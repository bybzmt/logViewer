<script>
    import Layout from "./lib/layout.svelte";
    import { getContext, onMount, onDestroy } from "svelte";

    let axios = getContext("axios");
    let rows = [];
    let selected = {};

    let ws;
    let msgs = [];

    let starttime = "";
    let endtime = "";
    let contains = "";
    let limit = "";

    let match = {
        Files: [],
        StartTime: 0,
        EndTime: 0,
        Limit: 0,
        BufSize: 1024 * 16,
    };

    function viewLogs() {
        axios({ url: "/api/viewLogs" }).then((resp) => {
            rows = resp.data;
        });
    }

    function search() {
        msgs = [];

        ws = new WebSocket("ws://" + API_BASE + "/api/logs");
        ws.onopen = () => {
            match.Files = [];
            rows.forEach((row) => {
                if (row.allfile) {
                    row.allfile.forEach((file) => {
                        match.Files.push({
                            Name: file,
                            TimeRegex: row.TimeRegex,
                            TimeLayout: row.TimeLayout,
                            Contains: [[contains]],
                        });
                    });
                }
            });

            match.Starttime = Date.parse(starttime) / 1000;
            match.EndTime = Date.parse(endtime) / 1000;
            match.Limit = parseInt(limit);

            if (isNaN(match.Starttime)) {
                match.Starttime = 0;
            }
            if (isNaN(match.EndTime)) {
                match.EndTime = parseInt(new Date().getTime() / 1000);
            }
            if (isNaN(match.Limit)) {
                match.Limit = 0;
            }

            console.log(match);

            ws.send(JSON.stringify(match));
        };

        ws.onclose = () => {
            console.log("close");
            ws = null;
        };
        ws.onerror = () => {
            console.log("error");
        };
        ws.onmessage = (evt) => {
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

    function listFiles(log) {
        return () => {
            if (!log.allfile) {
                log.allfile = [];
            }

            let files = log.Files.split(/\r\n|\r|\n/);
            files.forEach((str) => {
                str = str.trim();
                if (str == "") {
                    return;
                }

                axios({ url: "/api/glob", params: { pattern: str } }).then((resp) => {
                    if (resp.data.Err == "") {
                        log.allfile = log.allfile.concat(resp.data.Data);
                        //console.log(log.allfile);
                        rows = rows;
                    }
                });
            });
        };
    }

    onMount(() => {
        viewLogs();
    });
</script>

<Layout>
    <div class="flex">
        <div class="border w-1/5">
            {#each rows as log}
                <div>
                    <span on:click={listFiles(log)}>
                        {log.Note ? log.Note : log.ID}
                    </span>
                    <ul>
                        {#if log.allfile}
                            {#each log.allfile as file}
                                <li>{file}</li>
                            {/each}
                        {/if}
                    </ul>
                </div>
            {/each}
        </div>
        <div class="border w-full">
            <div>
                <input class="border w-[10em]" bind:value={starttime} placeholder="开始时间" />
                <input class="border w-[10em]" bind:value={endtime} placeholder="结束时间" />
                <select>
                    <option>开头</option>
                    <option>未尾</option>
                </select>
                <input class="border w-[5em]" bind:value={limit} placeholder="显示数量" />
                <input class="border w-[5em]" bind:value={contains} placeholder="匹配" />
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
