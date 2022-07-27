<script context="module">
    export const load = async ({ query, context }) => {
        let axios = context.get("axios");

        let rows = await axios({ url: "/api/viewLogs?sid=" + query.get("sid") }).then((resp) => {
            return resp.data;
        });

        return {
            props: {
                logs: rows,
                server_id: query.get("sid"),
            },
        };
    };
</script>

<script>
    import Layout from "./lib/layout.svelte";
    import { getContext, onMount, onDestroy } from "svelte";
    import { DateInput } from "date-picker-svelte";

    export let logs;
    export let server_id;

    let axios = getContext("axios");

    let ws;
    let ws_tick = 0;
    let ws_timer;
    let ws_onMsg;
    let msgs = [];
    let msgMax = 200;

    let starttime = new Date();
    let contains = "";
    let limit = "";
    let mode;

    starttime.setSeconds(0);
    starttime.setMinutes(0);
    starttime.setHours(0);

    let endtime = new Date();
    endtime.setTime(starttime.getTime() + 60 * 60 * 24 * 1000);

    async function getMatch(ws) {
        let match = {
            Files: [],
            StartTime: 0,
            EndTime: 0,
            Limit: 0,
            BufSize: 1024 * 512,
        };

        for (let i = 0; i < logs.length; i++) {
            if (logs[i].selected) {
                let data = await listFiles(logs[i], ws);
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

        match.StartTime = parseInt(starttime.getTime() / 1000);
        match.EndTime = parseInt(endtime.getTime() / 1000);
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

        return match;
    }

    function ws_grep(ws, match) {
        ws.send(JSON.stringify({ Op: "grep", Data: match }));
        ws.send(JSON.stringify({ Op: "read" }));

        ws_onMsg = function (res) {
            msgs.push(res.Data);

            if (msgs.length > msgMax) {
                msgs = msgs.slice(-msgMax);
            } else {
                msgs = msgs;
            }

            ws.send(JSON.stringify({ Op: "read" }));
        };
    }

    async function search() {
        msgs = [];

        ws = new WebSocket("ws://" + API_BASE + "/api/logs");
        ws.onopen = async () => {
            let match = await getMatch(ws);

            ws_grep(ws, match);
        };

        ws.onclose = () => {
            console.log("close");
            ws = null;
        };
        ws.onerror = () => {
            console.log("error");
        };
        ws.onmessage = async (evt) => {
            ws_tick = 0;

            let res;
            if (evt.data instanceof Blob) {
                let data = await evt.data.text();
                res = {Err:"", Data:data};
            } else {
                res = JSON.parse(evt.data);
            }

            if (res.Err != "") {
                ws.send(JSON.stringify({ Op: "close" }));
                msgs.push(res.Err);
                cancel();
                return;
            }

            ws_onMsg(res);
        };

        ws_tick = 0;
        ws_timer = setInterval(() => {
            if (ws_tick > 4) {
                msgs.push("Error: Timeout");
                cancel();
            } else if (ws_tick > 1) {
                ws.send(JSON.stringify({ Op: "ping" }));
            }
            ws_tick++;
        }, 1000);
    }

    function cancel() {
        clearInterval(ws_timer);
        ws.close();
    }

    async function listFiles(log, ws) {
        let files = log.Files.split(/\r\n|\r|\n/);
        let all = [];

        for (let i = 0; i < files.length; i++) {
            let str = files[i].trim();
            if (str != "") {
                const data = await new Promise((resolve, reject) => {
                    ws.send(JSON.stringify({ Op: "glob", Data: str }));

                    ws_onMsg = function (res) {
                        if (res.Err != "") {
                            throw new Error(res.Err);
                        } else {
                            resolve(res.Files);
                        }
                    };
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

    onMount(() => {});
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
            <div>
                <a href="#/logs?sid={server_id}">Change</a>
            </div>
        </div>
        <div class="border w-4/5">
            <div class="flex">
                <select bind:value={mode}>
                    <option value="1">普通模式</option>
                    <option value="2">未尾N行</option>
                </select>
                {#if mode == "1"}
                    <DateInput bind:value={starttime} />
                    <DateInput bind:value={endtime} />
                {/if}
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
    :root {
        --date-input-width: 10.5em;
    }
</style>
