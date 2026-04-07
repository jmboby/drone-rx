<script lang="ts">
	import '../app.css';

	let { children } = $props();
</script>

<div class="min-h-screen bg-navy-950 grid-bg">
	{@render children()}
</div>
