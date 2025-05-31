document.getElementById("uploadForm").addEventListener("submit", function (event) {
  const fileInput = document.getElementById("musicFile");
  const files = fileInput.files;
  const submitButton = document.getElementById("submitButton");
  const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB por archivo

  submitButton.disabled = true;
  submitButton.innerText = "Cargando...";

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

  for (let i = 0; i < files.length; i++) {
    const file = files[i];

    if (file.type !== "audio/mpeg") {
      event.preventDefault();
      submitButton.disabled = false;
      submitButton.innerText = "Upload";
      Swal.fire({
        icon: "error",
        title: "Archivo inválido",
        text: `El archivo "${file.name}" no es un MP3 válido.`,
      });
      return;
    }

    if (file.size > MAX_FILE_SIZE) {
      event.preventDefault();
      submitButton.disabled = false;
      submitButton.innerText = "Upload";
      Swal.fire({
        icon: "error",
        title: "Archivo demasiado grande",
        text: `El archivo "${file.name}" excede el tamaño máximo de 10MB.`,
      });
      return;
    }
  }

  // ✅ Si pasa todas las validaciones, continúa el envío
});



