/* global EasyMDE */

document.addEventListener("DOMContentLoaded", () => {
  const el = document.querySelector('textarea[name="desc"]');
  const cacheKeyInput = document.querySelector('[name="easymde_cache_key"]');

  if (!el || !cacheKeyInput) {
    console.error("Required elements not found");
    return;
  }

  new EasyMDE({
    element: el,
    autoDownloadFontAwesome: false,
    autosave: {
      enabled: true,
      delay: 1000,
      uniqueId: cacheKeyInput.value,
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
    previewRender: function (_, preview) {
      const form = document.getElementById("edit-form");
      if (!form) {
        console.error("Unable to find edit-form");
        return "Preview error";
      }

      (async function () {
        try {
          const response = await fetch("/preview", {
            method: "POST",
            headers: {
              "Content-Type": "application/x-www-form-urlencoded",
            },
            body: new URLSearchParams(new FormData(form)),
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
