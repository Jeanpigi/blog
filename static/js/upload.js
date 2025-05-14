document.getElementById("uploadForm").addEventListener("submit", function (event) {
  const fileInput = document.getElementById("musicFile");
  const files = fileInput.files;

  const submitButton = document.getElementById("submitButton");
  submitButton.disabled = true;
  submitButton.innerText = "Cargando...";

  // Verifica que haya al menos un archivo
  if (files.length === 0) {
    event.preventDefault();
    submitButton.disabled = false;
    submitButton.innerText = "Upload";

    Swal.fire({
      icon: "warning",
      title: "No Files",
      text: "Por favor selecciona al menos un archivo.",
    });
    return;
  }

  // Verifica que todos los archivos sean MP3
  for (let i = 0; i < files.length; i++) {
    if (files[i].type !== "audio/mpeg") {
      event.preventDefault();
      submitButton.disabled = false;
      submitButton.innerText = "Upload";

      Swal.fire({
        icon: "error",
        title: "Archivo inválido",
        text: `El archivo "${files[i].name}" no es un MP3 válido.`,
      });
      return;
    }
  }

  // ✅ Si pasa todas las validaciones, permite que se envíe el formulario
});



