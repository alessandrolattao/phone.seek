<script setup>
import { computed } from "vue";

const props = defineProps({
  phone: { type: Object, required: true },
  rank: { type: Number, default: 0 },
  maxScore: { type: Number, default: 1 },
  minScore: { type: Number, default: 0 },
});

const imageUrl = computed(() => {
  if (props.phone.image_file) {
    return `/api/images/${props.phone.image_file}`;
  }
  return props.phone.image_url || "";
});

const brandName = computed(() => {
  const b = props.phone.brand || "";
  return b.charAt(0).toUpperCase() + b.slice(1);
});

// Normalize score to 0-100 range within the result set
const relevanceWidth = computed(() => {
  if (!props.phone.score || !props.maxScore) return 0;
  const range = props.maxScore - props.minScore;
  if (range <= 0) return 100;
  // Map to 30-100% so even the worst result has visible bar
  return 30 + ((props.phone.score - props.minScore) / range) * 70;
});

const isTopMatch = computed(() => props.rank === 1);
</script>

<template>
  <div class="card bg-base-100/80 backdrop-blur-sm shadow-md hover:shadow-xl transition-all relative overflow-hidden group">
    <!-- Top match badge -->
    <div v-if="isTopMatch" class="absolute top-2.5 right-2 z-10">
      <span class="badge badge-primary badge-sm gap-1 font-mono text-[10px]">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-2.5 w-2.5" viewBox="0 0 20 20" fill="currentColor">
          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/>
        </svg>
        Top match
      </span>
    </div>

    <figure v-if="imageUrl" class="px-4 pt-4">
      <img
        :src="imageUrl"
        :alt="brandName + ' ' + phone.model"
        class="h-44 object-contain drop-shadow-sm group-hover:scale-[1.02] transition-transform duration-300"
        loading="lazy"
        @error="$event.target.style.display = 'none'"
      />
    </figure>

    <div class="card-body p-4 pt-3 gap-2">
      <h2 class="card-title text-sm leading-tight">
        <span class="font-bold">{{ brandName }}</span>
        <span class="font-normal text-base-content/70">{{ phone.model }}</span>
      </h2>

      <div class="grid grid-cols-1 gap-1 text-xs text-base-content/60 mt-1">
        <div v-if="phone.display" class="flex items-start gap-1.5">
          <svg class="h-3 w-3 mt-0.5 shrink-0 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <rect x="5" y="2" width="14" height="20" rx="2"/>
            <line x1="12" y1="18" x2="12" y2="18.01" stroke-width="3" stroke-linecap="round"/>
          </svg>
          <span>{{ phone.display }} <span v-if="phone.screen_size" class="text-base-content/40">&middot; {{ phone.screen_size.split(',')[0] }}</span></span>
        </div>

        <div v-if="phone.chipset" class="flex items-start gap-1.5">
          <svg class="h-3 w-3 mt-0.5 shrink-0 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <rect x="4" y="4" width="16" height="16" rx="2"/>
            <rect x="9" y="9" width="6" height="6"/>
            <line x1="9" y1="1" x2="9" y2="4"/><line x1="15" y1="1" x2="15" y2="4"/>
            <line x1="9" y1="20" x2="9" y2="23"/><line x1="15" y1="20" x2="15" y2="23"/>
            <line x1="20" y1="9" x2="23" y2="9"/><line x1="20" y1="14" x2="23" y2="14"/>
            <line x1="1" y1="9" x2="4" y2="9"/><line x1="1" y1="14" x2="4" y2="14"/>
          </svg>
          <span>{{ phone.chipset }}</span>
        </div>

        <div v-if="phone.storage" class="flex items-start gap-1.5">
          <svg class="h-3 w-3 mt-0.5 shrink-0 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path d="M4 7v10c0 2 1 3 3 3h10c2 0 3-1 3-3V7c0-2-1-3-3-3H7C5 4 4 5 4 7z"/>
            <path d="M16 4v4H8V4"/><path d="M8 20v-6h8v6"/><line x1="12" y1="16" x2="12" y2="16.01" stroke-width="3" stroke-linecap="round"/>
          </svg>
          <span>{{ phone.storage }}</span>
        </div>

        <div v-if="phone.battery" class="flex items-start gap-1.5">
          <svg class="h-3 w-3 mt-0.5 shrink-0 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <rect x="6" y="4" width="12" height="18" rx="2"/><line x1="10" y1="1" x2="14" y2="1"/>
            <path d="M10 14l2-4 2 4" fill="none"/>
          </svg>
          <span>{{ phone.battery.split(',')[0] }}</span>
        </div>

        <div v-if="phone.os" class="flex items-start gap-1.5">
          <svg class="h-3 w-3 mt-0.5 shrink-0 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="3"/><path d="M12 1v2m0 18v2m-9-11h2m18 0h2m-3.3-6.7-1.4 1.4m-11.6 11.6-1.4 1.4m0-14.4 1.4 1.4m11.6 11.6 1.4 1.4"/>
          </svg>
          <span class="truncate">{{ phone.os }}</span>
        </div>
      </div>

      <div v-if="phone.price" class="mt-auto pt-2 border-t border-base-content/5">
        <span class="text-xs font-semibold text-primary">{{ phone.price }}</span>
      </div>

      <div v-if="phone.score" class="pt-2 border-t border-base-content/5 flex items-center justify-between gap-2">
        <span class="text-[10px] font-mono text-base-content/35 uppercase tracking-wider">cosine similarity</span>
        <div class="flex items-center gap-1.5">
          <div class="w-12 h-1 bg-base-300 rounded-full overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-500"
              :class="isTopMatch ? 'bg-primary' : 'bg-primary/60'"
              :style="{ width: relevanceWidth + '%' }"
            />
          </div>
          <span class="text-[11px] font-mono tabular-nums" :class="isTopMatch ? 'text-primary font-semibold' : 'text-base-content/60'">{{ phone.score.toFixed(4) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
