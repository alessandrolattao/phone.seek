<script setup>
import { ref, onMounted, onUnmounted } from "vue";

const canvas = ref(null);
let animationId = null;

onMounted(() => {
  const ctx = canvas.value.getContext("2d");
  let w, h;

  function resize() {
    w = canvas.value.width = window.innerWidth;
    h = canvas.value.height = window.innerHeight;
  }

  resize();
  window.addEventListener("resize", resize);

  const waves = [
    { amp: 40, freq: 0.004, speed: 0.0012, x: 0.55, color1: "rgba(180, 60, 220, 0.4)", color2: "rgba(100, 40, 200, 0.1)" },
    { amp: 30, freq: 0.006, speed: -0.0008, x: 0.5, color1: "rgba(60, 140, 255, 0.35)", color2: "rgba(30, 80, 200, 0.05)" },
    { amp: 50, freq: 0.003, speed: 0.0005, x: 0.65, color1: "rgba(0, 210, 255, 0.3)", color2: "rgba(0, 180, 255, 0.05)" },
    { amp: 25, freq: 0.007, speed: -0.0015, x: 0.45, color1: "rgba(140, 80, 255, 0.25)", color2: "rgba(80, 40, 180, 0.05)" },
    { amp: 35, freq: 0.005, speed: 0.0003, x: 0.7, color1: "rgba(0, 240, 220, 0.2)", color2: "rgba(0, 200, 200, 0.02)" },
  ];

  let t = 0;

  function draw() {
    ctx.clearRect(0, 0, w, h);

    for (const wave of waves) {
      ctx.beginPath();
      ctx.moveTo(w, 0);

      for (let y = 0; y <= h; y += 3) {
        const x = wave.x * w + Math.sin(y * wave.freq + t * wave.speed * 60) * wave.amp
          + Math.sin(y * wave.freq * 0.5 + t * wave.speed * 30) * wave.amp * 0.5;
        ctx.lineTo(x, y);
      }

      ctx.lineTo(w, h);
      ctx.closePath();

      const grad = ctx.createLinearGradient(wave.x * w - wave.amp, 0, w, 0);
      grad.addColorStop(0, wave.color1);
      grad.addColorStop(1, wave.color2);
      ctx.fillStyle = grad;
      ctx.fill();
    }

    t++;
    animationId = requestAnimationFrame(draw);
  }

  draw();

  onUnmounted(() => {
    cancelAnimationFrame(animationId);
    window.removeEventListener("resize", resize);
  });
});
</script>

<template>
  <canvas ref="canvas" class="fixed inset-0 w-full h-full pointer-events-none z-0" />
</template>
