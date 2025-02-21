/* global EasyMDE */

document.addEventListener("DOMContentLoaded", () => {
  const el = document.querySelector('textarea[name="desc"]');
  const eventIdInput = document.querySelector('[name="event_id"]');

  if (!el || !eventIdInput) {
    console.error("Required elements not found");
    return;
  }

  new EasyMDE({
    element: el,
    autoDownloadFontAwesome: false,
    autosave: {
      enabled: true,
      delay: 1000,
      uniqueId: `desc-${eventIdInput.value}`,
    },
    forceSync: true,
    promptURLs: true,
    status: false,
    spellChecker: false,
    toolbar: [
      "bold",
      "italic",
      "strikethrough",
      "link",
      "|",
      "preview",
      "undo",
      "redo",
    ],
    previewRender: function (plainText, preview) {
      const csrfInput = document.querySelector('[name="csrf"]');
      const titleInput = document.querySelector('[name="title"]');
      if (!csrfInput || !titleInput) {
        console.error("Cannot find elements");
        return "Internal error";
      }

      (async function () {
        try {
          const response = await fetch("/preview", {
            method: "POST",
            headers: {
              "Content-Type": "application/x-www-form-urlencoded",
            },
            body: new URLSearchParams({
              csrf: csrfInput.value,
              title: titleInput.value,
              desc: plainText,
            }),
          });

          if (!response.ok) {
            throw new Error(
              `Preview failed: ${response.status} - ${response.statusText}`,
            );
          }

          const text = await response.text();

          preview.innerHTML = text;
        } catch (error) {
          console.error("Preview error:", error);
          preview.innerHTML = "Error loading preview";
        }
      })();
      return "Loading...";
    },
  });
});
