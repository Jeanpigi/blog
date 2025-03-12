document.addEventListener("DOMContentLoaded", () => {
  const usernameElement = document.getElementById("username");
  if (usernameElement) {
    const username = usernameElement.getAttribute("data-username");
    localStorage.setItem("username", username);
  } else {
    console.log("Elemento de nombre de usuario no encontrado en el DOM.");
  }

  if (document.getElementById("logoutButton")) {
    document.getElementById("logoutButton").addEventListener("click", () => {
      localStorage.clear();
      console.log("LocalStorage limpiado.");
    });
  } else {
    console.log("Botón de logout no encontrado.");
  }
});
