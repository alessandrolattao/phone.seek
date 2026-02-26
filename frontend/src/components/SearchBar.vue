<script setup>
import { ref } from "vue";

defineProps({ loading: Boolean });
const emit = defineEmits(["search-text", "search-image"]);

const query = ref("");
const dragOver = ref(false);
const cameraActive = ref(false);
const videoRef = ref(null);
let mediaStream = null;

function submitText() {
  if (query.value.trim()) {
    emit("search-text", query.value.trim());
  }
}

function handleFile(file) {
  if (file && file.type.startsWith("image/")) {
    emit("search-image", file);
  }
}

function onFileInput(e) {
  const file = e.target.files?.[0];
  if (file) handleFile(file);
}

function onDrop(e) {
  dragOver.value = false;
  const file = e.dataTransfer?.files?.[0];
  if (file) handleFile(file);
}

async function openCamera() {
  try {
    mediaStream = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: "environment" },
    });
    cameraActive.value = true;
    await new Promise((r) => setTimeout(r, 50));
    if (videoRef.value) {
      videoRef.value.srcObject = mediaStream;
    }
  } catch (err) {
    console.error("Camera access denied:", err);
  }
}

function capturePhoto() {
  const video = videoRef.value;
  if (!video) return;

  const canvas = document.createElement("canvas");
  canvas.width = video.videoWidth;
  canvas.height = video.videoHeight;
  canvas.getContext("2d").drawImage(video, 0, 0);

  canvas.toBlob((blob) => {
    if (blob) {
      const file = new File([blob], "camera-capture.jpg", { type: "image/jpeg" });
      closeCamera();
      handleFile(file);
    }
  }, "image/jpeg", 0.9);
}

function closeCamera() {
  if (mediaStream) {
    mediaStream.getTracks().forEach((t) => t.stop());
    mediaStream = null;
  }
  cameraActive.value = false;
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <form @submit.prevent="submitText" class="join w-full">
      <input
        v-model="query"
        type="text"
        placeholder="Search smartphones... (e.g. 'good camera phone under 500')"
        class="input join-item flex-1"
        :disabled="loading"
      />
      <button type="submit" class="btn btn-primary join-item" :disabled="loading || !query.trim()">
        Search
      </button>
    </form>

    <div class="flex gap-4">
      <div
        class="border-2 border-dashed rounded-xl p-6 text-center cursor-pointer transition-colors flex-1"
        :class="dragOver ? 'border-primary bg-primary/10' : 'border-base-content/20 hover:border-primary/50'"
        @dragover.prevent="dragOver = true"
        @dragleave="dragOver = false"
        @drop.prevent="onDrop"
        @click="$refs.fileInput.click()"
      >
        <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onFileInput" />
        <p class="text-base-content/60">
          Drop an image here or <span class="text-base-content underline underline-offset-2">click to upload</span>
        </p>
        <p class="text-xs text-base-content/40 mt-1">Search by photo</p>
      </div>

      <button
        class="btn btn-outline border-base-content/30 text-base-content hover:bg-base-content/10 hover:border-base-content/50 rounded-xl h-auto min-h-0 px-6"
        :disabled="loading || cameraActive"
        @click="openCamera"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        <span class="text-xs mt-1">Camera</span>
      </button>
    </div>

    <div v-if="cameraActive" class="relative rounded-xl overflow-hidden bg-black">
      <video ref="videoRef" autoplay playsinline class="w-full" />
      <div class="absolute bottom-4 left-0 right-0 flex justify-center gap-4">
        <button class="btn btn-circle btn-primary btn-lg" @click="capturePhoto">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <circle cx="12" cy="12" r="10" stroke-width="2" />
            <circle cx="12" cy="12" r="7" fill="currentColor" />
          </svg>
        </button>
        <button class="btn btn-circle btn-ghost btn-lg text-white" @click="closeCamera">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>
