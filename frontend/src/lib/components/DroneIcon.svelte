<script lang="ts">
	interface Props {
		size?: string;
		animated?: boolean;
	}

	let { size = 'w-6 h-6', animated = false }: Props = $props();
</script>

<svg
	class="{size} {animated ? 'animate-drone-fly' : ''}"
	viewBox="0 0 64 64"
	fill="none"
	stroke="currentColor"
	stroke-width="2.5"
	stroke-linecap="round"
	stroke-linejoin="round"
>
	<!-- Central body -->
	<rect x="26" y="26" width="12" height="12" rx="3" fill="currentColor" opacity="0.2" stroke="currentColor" />

	<!-- Arms -->
	<line x1="30" y1="28" x2="18" y2="16" />
	<line x1="34" y1="28" x2="46" y2="16" />
	<line x1="30" y1="36" x2="18" y2="48" />
	<line x1="34" y1="36" x2="46" y2="48" />

	<!-- Rotors -->
	<circle cx="18" cy="16" r="8" fill="currentColor" opacity="0.1" stroke="currentColor" stroke-width="2" />
	<circle cx="46" cy="16" r="8" fill="currentColor" opacity="0.1" stroke="currentColor" stroke-width="2" />
	<circle cx="18" cy="48" r="8" fill="currentColor" opacity="0.1" stroke="currentColor" stroke-width="2" />
	<circle cx="46" cy="48" r="8" fill="currentColor" opacity="0.1" stroke="currentColor" stroke-width="2" />

	<!-- Rotor centers -->
	<circle cx="18" cy="16" r="2" fill="currentColor" />
	<circle cx="46" cy="16" r="2" fill="currentColor" />
	<circle cx="18" cy="48" r="2" fill="currentColor" />
	<circle cx="46" cy="48" r="2" fill="currentColor" />
</svg>
