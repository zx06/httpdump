<script>
    let records = {};
    // fetch every 5 seconds
    function fetchRecords() {
        fetch("/api/record/")
            .then((response) => response.json())
            .then((data) => {
                records = data;
            });
    }
    fetchRecords();
    setInterval(() => {
        fetchRecords()
    }, 5000);
</script>

<ul>
    {#each Object.keys(records) as key}
        <li>id: {key},len: {records[key].length}</li>
    {/each}
</ul>
