document.addEventListener("DOMContentLoaded", () => {
  const nav = document.querySelector(".Nav-container");
  const btn = document.querySelector(".hamburger");
  const menu = nav?.querySelector("ul");

  if (!nav || !btn || !menu) return;

  const isMobile = () => window.matchMedia("(max-width: 768px)").matches;

  const closeMenu = () => {
    nav.classList.remove("is-active");
    btn.setAttribute("aria-expanded", "false");
  };

  btn.setAttribute("aria-expanded", "false");

  btn.addEventListener("click", () => {
    if (!isMobile()) return; // en PC no hace nada

    nav.classList.toggle("is-active");
    btn.setAttribute("aria-expanded", nav.classList.contains("is-active"));
  });

  // Si haces click en un link, cierra el menú
  menu.querySelectorAll("a").forEach((a) => {
    a.addEventListener("click", () => closeMenu());
  });

  // Si cambias el tamaño a desktop, asegura que quede cerrado
  window.addEventListener("resize", () => {
    if (!isMobile()) closeMenu();
  });
});
