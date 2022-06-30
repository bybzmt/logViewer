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

<Layout>index</Layout>

<style>
</style>
