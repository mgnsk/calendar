document.addEventListener("DOMContentLoaded", (event) => {
  const el = document.querySelector('textarea[name="desc"]');
  if (el) {
    const easyMDE = new EasyMDE({
      element: el,
      autoDownloadFontAwesome: false,
      autosave: {
        enabled: true,
        delay: 1000,
        uniqueId: "desc-" + document.querySelector('[name="event_id"]').value,
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
        (async function () {
          try {
            const response = await fetch("/preview", {
              method: "POST",
              headers: {
                "Content-Type": "application/x-www-form-urlencoded",
              },
              body: new URLSearchParams({
                csrf: document.querySelector('[name="csrf"]').value,
                title: document.querySelector('[name="title"]').value,
                desc: plainText,
              }),
            });

            if (!response.ok) {
              throw new Error(`Preview failed: ${response.status}`);
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
  }
});
