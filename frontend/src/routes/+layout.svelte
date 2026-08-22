<script lang="ts">
  import '../app.css';
  import Navbar from '$lib/components/Navbar.svelte';
  import Footer from '$lib/components/Footer.svelte';
</script>

<div class="min-h-screen flex flex-col bg-[#0b0c10] text-slate-100 selection:bg-rose-500 selection:text-white">
  <Navbar />
  <main class="flex-1">
    <slot />
  </main>
  <Footer />
</div>
