window.addEventListener("DOMContentLoaded", () => {
  const usernameElement = document.getElementById("username");
  if (usernameElement) {
    const username = usernameElement.getAttribute("data-username");
    localStorage.setItem("username", username);
    console.log(localStorage.getItem("username"));
  } else {
    console.log("Elemento de nombre de usuario no encontrado en el DOM.");
  }
});

document.addEventListener("DOMContentLoaded", function () {
  if (window.CKEDITOR) {
    CKEDITOR.replace("editor");
  } else {
    console.error("CKEDITOR no está definido.");
  }
});

document.getElementById("logoutButton").addEventListener("click", () => {
  // Limpiar localStorage
  localStorage.clear();
});
