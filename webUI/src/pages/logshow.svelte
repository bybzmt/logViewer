<script>
    import Layout from "./lib/layout.svelte";
    import { getContext, onMount, onDestroy } from "svelte";

    let axios = getContext("axios");
    let logs = [];
    let selected = {};

    let ws;
    let msgs = [];
    let msgMax = 200;

    let starttime = "";
    let endtime = "";
    let contains = "";
    let limit = "";

    function viewLogs() {
        axios({ url: "/api/viewLogs" }).then((resp) => {
            logs = resp.data;
        });
    }

    async function search() {
        let match = {
            Files: [],
            StartTime: 0,
            EndTime: 0,
            Limit: 0,
            BufSize: 1024 * 512,
        };

        for (let i = 0; i < logs.length; i++) {
            if (logs[i].selected) {
                let data = await listFiles(logs[i]);
                data.forEach((file) => {
                    match.Files.push({
                        Name: file,
                        TimeRegex: logs[i].TimeRegex,
                        TimeLayout: logs[i].TimeLayout,
                        Contains: contains ? [[contains]] : [],
                    });
                });
            }
        }

        if (match.Files.length < 1) {
            alert("没有匹配的日志");
            return;
        }

        match.StartTime = Date.parse(starttime) / 1000;
        match.EndTime = Date.parse(endtime) / 1000;
        match.Limit = parseInt(limit);

        if (isNaN(match.StartTime)) {
            match.StartTime = 0;
        }
        if (isNaN(match.EndTime)) {
            match.EndTime = parseInt(new Date().getTime() / 1000);
        }
        if (isNaN(match.Limit)) {
            match.Limit = 0;
        }

        console.log(match);

        msgs = [];

        ws = new WebSocket("ws://" + API_BASE + "/api/logs");
        ws.onopen = () => {
            ws.send(JSON.stringify({Op:"grep", "Data":match}));
            ws.send(JSON.stringify({Op:"read"}));
        };

        ws.onclose = () => {
            console.log("close");
            ws = null;
        };
        ws.onerror = () => {
            console.log("error");
        };
        ws.onmessage = async (evt) => {
            if (evt.data instanceof Blob) {
                msgs.push(await evt.data.text());

                ws.send(JSON.stringify({Op:"read"}));
            } else {
                msgs.push(evt.data);
                ws.send(JSON.stringify({Op:"close"}));
            }

            if (msgs.length > msgMax) {
                msgs = msgs.slice(-msgMax);
            } else {
                msgs = msgs;
            }
        };
    }

    function cancel() {
        ws.close();
    }

    async function listFiles(log) {
        let files = log.Files.split(/\r\n|\r|\n/);
        let all = [];

        for (let i = 0; i < files.length; i++) {
            let str = files[i].trim();
            if (str != "") {
                let data = await axios({ url: "/api/glob", params: { pattern: str } }).then((resp) => {
                    if (resp.data.Err == "") {
                        return resp.data.Data;
                    } else {
                        console.log(resp.data.Err);
                    }
                });

                if (data) {
                    all = all.concat(data);
                }
            }
        }

        return all;
    }

    function logSelect(log) {
        if (log.selected) {
            log.selected = false;
        } else {
            log.selected = true;
        }
        logs = logs;
    }

    onMount(() => {
        viewLogs();
    });
</script>

<Layout>
    <div class="flex">
        <div class="border w-1/5">
            {#each logs as log}
                <div>
                    <span on:click={logSelect(log)} class={log.selected ? "text-red-500" : ""}>
                        {log.Note ? log.Note : log.ID}
                    </span>
                </div>
            {/each}
        </div>
        <div class="border w-4/5">
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
                    <p class="border-2">{msg}</p>
                {/each}
            </div>
        </div>
    </div>
</Layout>

<style>
</style>
