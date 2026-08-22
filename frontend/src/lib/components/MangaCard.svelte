<script lang="ts">
  import type { Manga } from '$lib/types';
  import { Eye, Star, BookOpen } from 'lucide-svelte';

  export let manga: Manga;

  function formatViews(views: number): string {
    if (views >= 1_000_000) return (views / 1_000_000).toFixed(1) + 'M';
    if (views >= 1_000) return (views / 1_000).toFixed(1) + 'K';
    return views.toString();
  }
</script>

<a
  href={`/manga/${manga.slug || manga.id}`}
  class="group relative flex flex-col rounded-2xl overflow-hidden glass-card transition-all duration-300 hover:-translate-y-1 hover:shadow-xl hover:shadow-rose-950/20"
>
  <!-- Cover Image Container -->
  <div class="relative aspect-[3/4] w-full overflow-hidden bg-slate-900">
    {#if manga.coverUrl}
      <img
        src={manga.coverUrl}
        alt={manga.title}
        loading="lazy"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500 ease-out"
        onerror={(e) => {
          // @ts-ignore
          e.target.style.display = 'none';
        }}
      />
    {:else}
      <div class="w-full h-full flex flex-col items-center justify-center text-slate-600">
        <BookOpen class="w-10 h-10 mb-2 stroke-1" />
        <span class="text-xs">No Cover</span>
      </div>
    {/if}

    <!-- Gradient Overlay -->
    <div class="absolute inset-0 bg-gradient-to-t from-slate-950 via-slate-950/20 to-transparent opacity-80 group-hover:opacity-90 transition-opacity"></div>

    <!-- Status Badge -->
    <div class="absolute top-2.5 left-2.5 flex items-center gap-1.5">
      <span class="px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider rounded-md backdrop-blur-md {manga.status === 'Completed' ? 'bg-emerald-500/80 text-white' : 'bg-rose-600/80 text-white'} shadow-sm">
        {manga.status || 'Ongoing'}
      </span>
    </div>

    <!-- Rating Badge -->
    {#if manga.rating}
      <div class="absolute top-2.5 right-2.5 flex items-center gap-1 px-2 py-0.5 text-[10px] font-bold rounded-md bg-slate-900/80 text-amber-300 backdrop-blur-md border border-white/10">
        <Star class="w-3 h-3 fill-amber-400 text-amber-400" />
        <span>{manga.rating.toFixed(1)}</span>
      </div>
    {/if}

    <!-- Bottom info on image -->
    <div class="absolute bottom-2.5 left-2.5 right-2.5 flex items-center justify-between text-[11px] text-slate-300">
      {#if manga.lastChapterNumber !== undefined && manga.lastChapterNumber !== null}
        <span class="font-bold text-rose-400 bg-slate-900/80 px-2 py-0.5 rounded-md border border-white/10">
          Ch. {manga.lastChapterNumber}
        </span>
      {:else}
        <span class="text-slate-400 text-[10px]">Chưa có chapter</span>
      {/if}

      <div class="flex items-center gap-1 text-slate-400">
        <Eye class="w-3 h-3" />
        <span>{formatViews(manga.views || 0)}</span>
      </div>
    </div>
  </div>

  <!-- Content / Info -->
  <div class="p-3.5 flex flex-col flex-1 justify-between bg-slate-900/40">
    <div>
      <h3 class="font-bold text-sm text-slate-100 line-clamp-1 group-hover:text-rose-400 transition-colors" title={manga.title}>
        {manga.title}
      </h3>
      {#if manga.author}
        <p class="text-[11px] text-slate-400 line-clamp-1 mt-0.5">{manga.author}</p>
      {/if}
    </div>

    <!-- Genres -->
    {#if manga.genres && manga.genres.length > 0}
      <div class="flex flex-wrap gap-1 mt-2.5">
        {#each manga.genres.slice(0, 2) as genre}
          <span class="text-[9px] font-medium px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700/50">
            {genre}
          </span>
        {/each}
        {#if manga.genres.length > 2}
          <span class="text-[9px] text-slate-500 px-1 py-0.5">+{manga.genres.length - 2}</span>
        {/if}
      </div>
    {/if}
  </div>
</a>
