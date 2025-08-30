document.addEventListener("DOMContentLoaded", async () => {
  const grid = document.getElementById("container-historias");
  if (!grid) return;

  Swal.fire({
    title: "Cargando…",
    text: "Buscando historias.",
    showConfirmButton: false,
    allowOutsideClick: false,
    didOpen: () => Swal.showLoading(),
  });

  try {
    const resp = await fetch("/api/histories", { cache: "no-store" });
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    const data = await resp.json();
    const posts = Array.isArray(data) ? data : [];
    render(posts);
  } catch (e) {
    console.error("[historias] error:", e);
    grid.innerHTML = `
      <div class="stories-empty">
        No pudimos cargar las historias.
        <button class="stories-btn" onclick="location.reload()">Reintentar</button>
      </div>`;
  } finally {
    Swal.close();
  }

  function render(posts) {
    grid.innerHTML = "";
    if (!posts.length) {
      grid.innerHTML = `<div class="stories-empty">Aún no hay historias 👀</div>`;
      return;
    }

    const frag = document.createDocumentFragment();
    posts.forEach(post => {
      const card = el("article", "post-card");

      // Head
      const head = el("header", "post-card__head");
      const emoji = el("span", "post-card__emoji", randomEmoji());
      const h2 = el("h2", "post-card__title");
      const link = el("a", "post-card__link", post.title || "Sin título");
      link.href = `/post/${post.id}`;
      link.rel = "prefetch";
      h2.appendChild(link);
      head.append(emoji, h2);

      // Meta
      const meta = el("div", "post-card__meta");
      const time = document.createElement("time");
      time.dateTime = post.created_at || "";
      time.textContent = formatDate(post.created_at);
      const read = el("span", "post-card__read", readingTime(post.content || post.description || ""));
      meta.append(time, read);

      // Body
      const body = el("div", "post-card__body");
      body.appendChild(el("p", null, post.description || "Sin descripción."));

      card.append(head, meta, body);
      frag.appendChild(card);
    });
    grid.appendChild(frag);
    revealOnView();
  }

  // ---- helpers ----
  function el(tag, className, text) {
    const n = document.createElement(tag);
    if (className) n.className = className;
    if (text) n.textContent = text;
    return n;
  }

  function revealOnView() {
    const cards = document.querySelectorAll(".post-card");
    const io = new IntersectionObserver((entries, obs) => {
      entries.forEach(en => {
        if (en.isIntersecting) { en.target.classList.add("is-visible"); obs.unobserve(en.target); }
      });
    }, { threshold: 0.1 });
    cards.forEach(c => io.observe(c));
  }

  function readingTime(text) {
    const words = (text || "").trim().split(/\s+/).filter(Boolean).length;
    const mins = Math.max(1, Math.ceil(words / 200));
    return `⏱️ ${mins} min de lectura`;
  }

  function formatDate(iso) {
    if (!iso) return "Fecha desconocida";
    try {
      return new Date(iso).toLocaleDateString("es-ES", { year: "numeric", month: "long", day: "numeric" });
    } catch { return "Fecha desconocida"; }
  }

  function randomEmoji() {
    const emojis = ["📖","🎬","🧩","🌙","🔥","🚪","🕯️","🧠","🕵️","🦊"];
    return emojis[Math.floor(Math.random() * emojis.length)];
  }
});

/* --- Card clicable + prefetch para navegación rápida --- */
document.addEventListener("click", (e) => {
  const card = e.target.closest(".post-card");
  if (!card) return;
  if (e.target.closest("a")) return; // ya es el link
  const link = card.querySelector(".post-card__link");
  if (!link) return;
  e.preventDefault();
  window.location.href = link.href;
});

document.addEventListener("keydown", (e) => {
  if (e.key !== "Enter" && e.key !== " ") return;
  const card = document.activeElement.closest?.(".post-card");
  if (!card) return;
  const link = card.querySelector(".post-card__link");
  if (!link) return;
  e.preventDefault();
  window.location.href = link.href;
});

const prefetch = (url) => {
  try {
    const l = document.createElement("link");
    l.rel = "prefetch";
    l.href = url;
    l.as = "document";
    document.head.appendChild(l);
  } catch {}
};

["mouseover","touchstart"].forEach(ev => {
  document.addEventListener(ev, (e) => {
    const card = e.target.closest(".post-card");
    if (!card) return;
    const link = card.querySelector(".post-card__link");
    if (link?.href) prefetch(link.href);
  });
});
