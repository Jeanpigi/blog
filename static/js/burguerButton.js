document.addEventListener("DOMContentLoaded", () => {
  const nav = document.querySelector(".Nav-container");
  const btn = document.querySelector(".hamburger");
  const menu = nav?.querySelector("ul");

  if (!nav || !btn || !menu) return;

  const isMobile = () => window.matchMedia("(max-width: 768px)").matches;

  const openMenu = () => {
    nav.classList.add("is-active");
    btn.setAttribute("aria-expanded", "true");
    btn.setAttribute("aria-label", "Cerrar menú");
    btn.innerHTML = "&#10005;"; // X cuando está abierto
  };

  const closeMenu = () => {
    nav.classList.remove("is-active");
    btn.setAttribute("aria-expanded", "false");
    btn.setAttribute("aria-label", "Abrir menú");
    btn.innerHTML = "&#9776;"; // ☰ cuando está cerrado
  };

  btn.addEventListener("click", (e) => {
    e.stopPropagation();
    if (!isMobile()) return;
    nav.classList.contains("is-active") ? closeMenu() : openMenu();
  });

  // Cerrar al hacer click en un enlace
  menu.querySelectorAll("a").forEach((a) => {
    a.addEventListener("click", () => closeMenu());
  });

  // Cerrar al hacer click fuera del menú
  document.addEventListener("click", (e) => {
    if (isMobile() && nav.classList.contains("is-active") && !nav.contains(e.target)) {
      closeMenu();
    }
  });

  // Cerrar al redimensionar a desktop
  window.addEventListener("resize", () => {
    if (!isMobile()) closeMenu();
  });
});
