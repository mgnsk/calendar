/* global EasyMDE */
/* global GeoSearch */
/* global $ */

document.addEventListener("DOMContentLoaded", () => {
  const el = document.querySelector('textarea[name="desc"]');
  const cacheKeyInput = document.querySelector('[name="easymde_cache_key"]');

  if (!el || !cacheKeyInput) {
    return;
  }

  new EasyMDE({
    element: el,
    autoDownloadFontAwesome: false,
    autosave: {
      enabled: false,
      // delay: 1000,
      // uniqueId: cacheKeyInput.value,
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

          if (!response.ok || response.status !== 200) {
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

  $('[name="location"]').autocomplete({
    source: function (request, response) {
      const query = request.term.trim();
      if (!query) {
        response([]);
        return;
      }

      $("#location-spinner").css("opacity", "1");

      const providerform = new GeoSearch.OpenStreetMapProvider({
        params: {
          limit: 5,
          "accept-language": "en",
        },
      });
      return providerform
        .search({ query })
        .then(function (results) {
          response(results);
        })
        .catch(function (error) {
          response([]);
          alert(`Location search failed: ${error}`);
        })
        .finally(function () {
          $("#location-spinner").css("opacity", "0");
        });
    },
    select: async function (_, ui) {
      $("#location-spinner").css("opacity", "1");

      const lat = ui.item.y;
      const long = ui.item.x;

      $('[name="latitude"]').val(lat);
      $('[name="longitude"]').val(long);

      const params = new URLSearchParams();
      params.set("latitude", lat);
      params.set("longitude", long);
      params.set(
        "user_timezone",
        Intl.DateTimeFormat().resolvedOptions().timeZone,
      );

      try {
        const response = await fetch("/gettimezone?" + params.toString());

        if (!response.ok || response.status !== 200) {
          throw new Error(
            `Fetching timezone failed: ${response.status} - ${response.statusText}`,
          );
        }

        const json = await response.json();

        $('[name="timezone_offset"]').val(json["timezone_offset"]);
      } catch (error) {
        alert(`Fetching timezone failed: ${error}`);
      } finally {
        $("#location-spinner").css("opacity", "0");
      }
    },
    delay: 1000,
    minLength: 3,
  });
});
