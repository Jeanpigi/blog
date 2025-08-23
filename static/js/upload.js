// /static/js/upload.js
document.getElementById("uploadForm").addEventListener("submit", async function (e) {
  e.preventDefault();

  const fileInput = document.getElementById("musicFile"); 
  const submitButton = document.getElementById("submitButton"); 
  const files = fileInput.files;
  const MAX_FILE_SIZE = 20 * 1024 * 1024; // 20MB por archivo

  // UI
  submitButton.disabled = true;
  submitButton.innerText = "Subiendo...";

  // Validaciones básicas
  if (files.length === 0) {
    submitButton.disabled = false;
    submitButton.innerText = "Upload";
    return Swal.fire("Sin archivos", "Selecciona al menos un archivo.", "warning");
  }

  if (files.length > 5) {
    submitButton.disabled = false;
    submitButton.innerText = "Upload";
    return Swal.fire("Demasiados archivos", "Solo puedes subir hasta 5 a la vez.", "error");
  }

  // Validaciones por archivo (rápidas en el cliente)
  for (const f of files) {
    const name = f.name.toLowerCase();
    const isMp3 = name.endsWith(".mp3") && f.type === "audio/mpeg";
    if (!isMp3) {
      submitButton.disabled = false;
      submitButton.innerText = "Upload";
      return Swal.fire("Archivo inválido", `${f.name} no es un MP3 válido.`, "error");
    }
    if (f.size > MAX_FILE_SIZE) {
      submitButton.disabled = false;
      submitButton.innerText = "Upload";
      return Swal.fire("Archivo demasiado grande", `${f.name} excede los 20MB.`, "error");
    }
  }

  // Armar payload
  const formData = new FormData();
  for (const f of files) formData.append("musicFiles", f); // <- nombre del campo que espera tu backend

  try {
    const res = await fetch("/radio/upload", { method: "POST", body: formData });
    // Manejo de errores de red/HTTP
    if (!res.ok) {
      const msg = await res.text();
      throw new Error(msg || "Error al subir los archivos");
    }

    const data = await res.json(); // { uploaded: string[], skipped: [{file, reason}] }
    const uploaded = Array.isArray(data.uploaded) ? data.uploaded : [];
    const skipped  = Array.isArray(data.skipped)  ? data.skipped  : [];

    // 👉 Mostrar resumen compacto: solo detalles de omitidos, conteos arriba
    await Swal.fire({
      title: "Subida completada",
      html: `
        ✅ Archivos subidos: ${uploaded.length}<br>
        ⚠️ Omitidos: ${skipped.length}
        ${skipped.length > 0 ? `<hr>${skipped.map(s => `• ${s.file}: ${s.reason}`).join("<br>")}` : ""}
      `,
      icon: skipped.length > 0 ? "info" : "success",
      confirmButtonText: "OK"
    });
  } catch (err) {
    console.error(err);
    await Swal.fire("Error", err.message, "error");
  } finally {
    // Reset UI
    submitButton.disabled = false;
    submitButton.innerText = "Upload";
    fileInput.value = ""; // limpia la selección
  }
});


