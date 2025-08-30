document.addEventListener("DOMContentLoaded", async () => {
  const container = document.getElementById("container-tecnologia");
  if (!container) return;

  // Limpia cualquier HTML previo (por si otro script ya pintó)
  container.innerHTML = "";

  Swal.fire({
    title: "Cargando…",
    text: "Trayendo posts de tecnología.",
    showConfirmButton: false,
    allowOutsideClick: false,
    didOpen: () => Swal.showLoading(),
  });

  try {
    const resp = await fetch("/api/categories", { cache: "no-store" });
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    const data = await resp.json();
    const posts = Array.isArray(data) ? data : [];

    render(posts);
    console.log("[tech] posts renderizados:", posts.length);

  } catch (e) {
    console.error("[tech] error:", e);
    container.innerHTML = `<div class="tech-empty">No pudimos cargar los posts. <button class="tech-btn" onclick="location.reload()">Reintentar</button></div>`;
  } finally {
    Swal.close();
  }

  function render(posts) {
    container.innerHTML = "";
    if (!posts.length) {
      container.innerHTML = `<div class="tech-empty">No hay posts por ahora 👀</div>`;
      return;
    }

    const frag = document.createDocumentFragment();

    posts.forEach((post) => {
      const card = document.createElement("article");
      // Fuerza la clase moderna
      card.className = "post-card";

      // HEAD
      const head = el("header", "post-card__head");
      const emoji = el("span", "post-card__emoji", randomEmoji());
      const h2 = el("h2", "post-card__title");
      const link = el("a", "post-card__link", post.title || "Sin título");
      link.href = `/post/${post.id}`;
      link.rel = "prefetch";

      h2.appendChild(link);
      head.append(emoji, h2);

      // META
      const meta = el("div", "post-card__meta");
      const time = document.createElement("time");
      time.dateTime = post.created_at || "";
      time.textContent = formatDate(post.created_at);
      const read = el("span", "post-card__read", readingTime(post.content || post.description || ""));
      meta.append(time, read);

      // BODY
      const body = el("div", "post-card__body");
      body.appendChild(el("p", null, post.description || "Sin descripción."));

      card.append(head, meta, body);
      frag.appendChild(card);
    });

    container.appendChild(frag);
    revealOnView();
  }

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
    const emojis = ["💡","🧠","⚙️","🛠️","🚀","🔐","🖥️","📦","🧪","🧰"];
    return emojis[Math.floor(Math.random() * emojis.length)];
  }
});
