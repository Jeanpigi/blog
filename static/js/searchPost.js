document.addEventListener("DOMContentLoaded", () => {
  const searchButton = document.getElementById("searchButton");
  if (searchButton) {
    searchButton.addEventListener("click", () => {
      let query = document.getElementById("searchQuery").value; // Mover la obtención del valor de 'query' dentro del evento 'click'
      fetch(`/api/posts/${query}`)
        .then((response) => response.json())
        .then((posts) => {
          console.log(posts);
          // Aquí deberías agregar el código para actualizar el DOM con los posts encontrados
        })
        .catch((error) => console.error("Error:", error)); // Este es el lugar correcto para .catch
    });
  } else {
    console.error("El botón de búsqueda no fue encontrado.");
  }
});
