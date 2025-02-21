/* global Mark */

const punct = ":;.,-–—‒_(){}[]!'\"+=".split("");

async function highlightResults(targetNode, searchValue) {
  return new Promise((resolve) => {
    const mark = new Mark(targetNode);
    const s = searchValue.replace(/"+/g, ""); // Remove double quotes.
    const results = [];

    mark.mark(s, {
      ignorePunctuation: punct,
      each: function (el) {
        results.push(el);
      },
      done: function () {
        resolve(results);
      },
    });
  });
}

/* exported setSearch */
function setSearch(s) {
  let el = document.getElementById("search");
  el.value = s;
}

document.addEventListener("DOMContentLoaded", () => {
  const eventList = document.getElementById("event-list");
  const search = document.getElementById("search");
  if (search && eventList) {
    eventList.addEventListener("htmx:afterSettle", async function (evt) {
      switch (evt.detail.target.id) {
        case "event-list":
          // Initial events loaded (on tab switch or search query change).
          await highlightResults(evt.detail.elt, search.value);
          window.scrollTo({ top: 0, behavior: "smooth" });
          break;

        case "load-more":
          // Infinite scroll loaded more events.
          // Highlight the added event.
          await highlightResults(evt.detail.elt, search.value);
          break;
      }
    });
  }
});
