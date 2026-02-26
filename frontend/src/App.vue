<script setup>
import { ref, watch, computed, onMounted } from "vue";
import PhoneCard from "./components/PhoneCard.vue";
import SearchBar from "./components/SearchBar.vue";
import SearchFilters from "./components/SearchFilters.vue";
import WaveBackground from "./components/WaveBackground.vue";

const results = ref([]);
const loading = ref(false);
const searched = ref(false);
const searchTime = ref(null);
const lastQuery = ref("");
const filterOptions = ref({ brands: [], nfc: [], network: [], os: [], display_type: [] });
const activeFilters = ref({ brand: "", nfc: "", network: "", os: "", display_type: "", price_min: "", price_max: "" });

const maxScore = computed(() => {
  if (!results.value.length) return 1;
  return Math.max(...results.value.map((r) => r.score || 0));
});

const minScore = computed(() => {
  if (!results.value.length) return 0;
  return Math.min(...results.value.map((r) => r.score || 0));
});

watch(activeFilters, () => {
  if (lastQuery.value) {
    onTextSearch(lastQuery.value);
  }
}, { deep: true });

onMounted(async () => {
  try {
    const res = await fetch("/api/filters");
    filterOptions.value = await res.json();
  } catch (err) {
    console.error("Failed to load filters:", err);
  }
});

function buildFilterParams() {
  const params = new URLSearchParams();
  if (activeFilters.value.brand) params.set("brand", activeFilters.value.brand);
  if (activeFilters.value.nfc) params.set("nfc", activeFilters.value.nfc);
  if (activeFilters.value.network) params.set("network", activeFilters.value.network);
  if (activeFilters.value.os) params.set("os", activeFilters.value.os);
  if (activeFilters.value.display_type) params.set("display_type", activeFilters.value.display_type);
  if (activeFilters.value.price_min) params.set("price_min", activeFilters.value.price_min);
  if (activeFilters.value.price_max) params.set("price_max", activeFilters.value.price_max);
  return params;
}

async function onTextSearch(query) {
  lastQuery.value = query;
  loading.value = true;
  searched.value = true;
  results.value = [];
  searchTime.value = null;

  try {
    const params = buildFilterParams();
    params.set("q", query);
    const res = await fetch(`/api/search?${params}`);
    const data = await res.json();
    results.value = data.results || [];
    searchTime.value = data.time_ms;
  } catch (err) {
    console.error("Search failed:", err);
  } finally {
    loading.value = false;
  }
}

async function onImageSearch(file) {
  loading.value = true;
  searched.value = true;
  results.value = [];
  searchTime.value = null;

  try {
    const formData = new FormData();
    formData.append("image", file);
    const filterParams = buildFilterParams();
    for (const [k, v] of filterParams) {
      formData.append(k, v);
    }
    const res = await fetch("/api/search/image", {
      method: "POST",
      body: formData,
    });
    const data = await res.json();
    results.value = data.results || [];
    searchTime.value = data.time_ms;
  } catch (err) {
    console.error("Image search failed:", err);
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="min-h-screen bg-base-300 relative">
    <WaveBackground />

    <div class="container mx-auto px-4 pt-10 pb-8 max-w-6xl relative z-10">
      <div class="text-center mb-6">
        <h1 class="text-4xl font-extrabold tracking-tight font-mono">
          <span class="text-primary">&gt;</span> phone<span class="text-primary">.</span>seek<span class="text-primary">()</span>
        </h1>
        <p class="text-base-content/50 mt-1 text-sm">Multimodal semantic search powered by CLIP + MiniLM embeddings on Qdrant</p>
      </div>

      <SearchFilters v-model="activeFilters" :options="filterOptions" class="justify-center" />

      <SearchBar @search-text="onTextSearch" @search-image="onImageSearch" :loading="loading" class="mt-3" />

      <div v-if="loading" class="flex justify-center py-16">
        <span class="loading loading-spinner loading-lg"></span>
      </div>

      <template v-else-if="results.length > 0">
        <p v-if="searchTime !== null" class="text-base-content/40 text-xs font-mono mt-5 mb-2">
          {{ results.length }} results in {{ searchTime }}ms
        </p>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          <PhoneCard
            v-for="(phone, i) in results"
            :key="i"
            :phone="phone"
            :rank="i + 1"
            :maxScore="maxScore"
            :minScore="minScore"
          />
        </div>
      </template>

      <div v-else-if="searched" class="text-center py-16">
        <p class="text-base-content/60 text-lg">No results found</p>
      </div>

      <div v-else class="text-center py-16">
        <p class="text-base-content/60 text-lg">Search smartphones by text or upload an image</p>
      </div>
    </div>
  </div>
</template>
