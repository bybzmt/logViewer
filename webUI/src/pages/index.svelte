<script>
    import Layout from "./lib/layout.svelte";
    import { onMount, onDestroy } from "svelte";

    let timer;
    onMount(() => {
        showStatus();

        timer = setInterval(function () {
            showStatus();
        }, 1000);
    });

    onDestroy(() => {
        clearInterval(timer);
    });
</script>

<Layout>
    <a>配置</a>
    <a>日志列表</a>
</Layout>

<style>
</style>
