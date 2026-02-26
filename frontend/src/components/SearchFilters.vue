<script setup>
import { computed } from "vue";

const props = defineProps({
  modelValue: Object,
  options: Object,
});
const emit = defineEmits(["update:modelValue"]);

const filters = computed({
  get: () => props.modelValue,
  set: (v) => emit("update:modelValue", v),
});

const hasActiveFilters = computed(() => {
  const f = props.modelValue;
  return f.brand || f.nfc || f.network || f.os || f.display_type || f.price_min || f.price_max;
});

function update(key, value) {
  emit("update:modelValue", { ...props.modelValue, [key]: value });
}

function clearAll() {
  emit("update:modelValue", {
    brand: "", nfc: "", network: "", os: "", display_type: "", price_min: "", price_max: "",
  });
}
</script>

<template>
  <div class="flex flex-wrap gap-2 items-center">
    <select
      class="select select-bordered select-sm select-xs min-w-0 w-auto"
      :value="filters.brand"
      @change="update('brand', $event.target.value)"
    >
      <option value="">Brand</option>
      <option v-for="b in options.brands" :key="b" :value="b">
        {{ b.charAt(0).toUpperCase() + b.slice(1) }}
      </option>
    </select>

    <select
      class="select select-bordered select-sm select-xs min-w-0 w-auto"
      :value="filters.os"
      @change="update('os', $event.target.value)"
    >
      <option value="">OS</option>
      <option v-for="o in options.os" :key="o" :value="o">{{ o }}</option>
    </select>

    <select
      class="select select-bordered select-sm select-xs min-w-0 w-auto"
      :value="filters.display_type"
      @change="update('display_type', $event.target.value)"
    >
      <option value="">Display</option>
      <option v-for="d in options.display_type" :key="d" :value="d">{{ d }}</option>
    </select>

    <select
      class="select select-bordered select-sm select-xs min-w-0 w-auto"
      :value="filters.nfc"
      @change="update('nfc', $event.target.value)"
    >
      <option value="">NFC</option>
      <option value="Yes">Yes</option>
      <option value="No">No</option>
    </select>

    <div class="flex gap-0.5">
      <button
        v-for="net in options.network"
        :key="net"
        class="btn btn-xs rounded-sm font-mono"
        :class="filters.network === net ? 'btn-primary' : 'btn-ghost text-base-content/50'"
        @click="update('network', filters.network === net ? '' : net)"
      >
        {{ net }}
      </button>
    </div>

    <div class="flex items-center gap-1 text-xs text-base-content/40">
      <span class="font-mono">EUR</span>
      <input
        type="number"
        class="input input-bordered input-xs w-16 text-center font-mono"
        placeholder="min"
        :value="filters.price_min"
        @change="update('price_min', $event.target.value)"
      />
      <span>-</span>
      <input
        type="number"
        class="input input-bordered input-xs w-16 text-center font-mono"
        placeholder="max"
        :value="filters.price_max"
        @change="update('price_max', $event.target.value)"
      />
    </div>

    <button
      v-if="hasActiveFilters"
      class="btn btn-xs btn-ghost text-error font-mono"
      @click="clearAll"
    >
      clear
    </button>
  </div>
</template>
