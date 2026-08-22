<script lang="ts">
  import { authStore } from '$lib/stores';
  import { Search, BookOpen, Shield, LogOut, Smartphone, Sparkles } from 'lucide-svelte';
  import { goto } from '$app/navigation';

  let searchQuery = '';

  function handleSearch(e: Event) {
    e.preventDefault();
    if (searchQuery.trim()) {
      goto(`/?q=${encodeURIComponent(searchQuery.trim())}`);
    }
  }

  function handleLogout() {
    authStore.logout();
    goto('/');
  }
</script>

<header class="sticky top-0 z-40 w-full glass border-b border-white/5">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-4">
    <!-- Brand Logo -->
    <a href="/" class="flex items-center gap-2.5 group shrink-0">
      <img src="/icon.png" alt="MangaHost" class="w-9 h-9 rounded-xl object-cover shadow-lg shadow-rose-500/20 group-hover:scale-105 transition-transform duration-200 border border-white/10" />
      <div class="flex flex-col">
        <span class="font-extrabold text-lg tracking-tight bg-gradient-to-r from-white via-slate-200 to-rose-400 bg-clip-text text-transparent">
          Manga<span class="text-rose-500">Host</span>
        </span>
        <span class="text-[10px] font-medium text-slate-400 -mt-1 tracking-wider uppercase flex items-center gap-1">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span> Cloudflare R2
        </span>
      </div>
    </a>

    <!-- Search Bar -->
    <form on:submit={handleSearch} class="flex-1 max-w-md hidden md:block">
      <div class="relative">
        <Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2 pointer-events-none" />
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Tìm tên truyện, tác giả..."
          class="w-full pl-10 pr-4 py-2 bg-slate-900/80 border border-slate-800 rounded-xl text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:border-rose-500/60 focus:ring-2 focus:ring-rose-500/10 transition-all"
        />
      </div>
    </form>

    <!-- Navigation actions -->
    <div class="flex items-center gap-3">
      <a
        href="/#browse"
        class="hidden sm:inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800/60 rounded-lg transition-colors"
      >
        <Sparkles class="w-3.5 h-3.5 text-amber-400" />
        Khám phá
      </a>

      <!-- Paperback info link -->
      <a
        href="https://github.com/Paperback-iOS/extensions-default"
        target="_blank"
        rel="noopener noreferrer"
        class="hidden sm:inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800/60 rounded-lg transition-colors"
        title="Paperback iOS App Integration"
      >
        <Smartphone class="w-3.5 h-3.5 text-blue-400" />
        Paperback iOS
      </a>

      {#if $authStore.isLoggedIn}
        <a
          href="/admin"
          class="inline-flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-bold text-white bg-rose-600 hover:bg-rose-500 shadow-md shadow-rose-600/20 rounded-xl transition-all"
        >
          <Shield class="w-3.5 h-3.5" />
          Admin Panel
        </a>
        <button
          on:click={handleLogout}
          class="p-2 text-slate-400 hover:text-rose-400 hover:bg-slate-800/60 rounded-xl transition-colors"
          title="Đăng xuất"
        >
          <LogOut class="w-4 h-4" />
        </button>
      {:else}
        <a
          href="/admin/login"
          class="inline-flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 rounded-xl transition-colors"
        >
          <Shield class="w-3.5 h-3.5 text-slate-400" />
          Admin
        </a>
      {/if}
    </div>
  </div>
</header>
