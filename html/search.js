const punct = ":;.,-–—‒_(){}[]!'\"+=".split("");

async function highlightResults(node) {
  return new Promise((resolve) => {
    const mark = new Mark(node);
    const s = search.value.replace(/"+/g, ""); // Remove double quotes.
    const results = [];

    mark.mark(s, {
      ignorePunctuation: punct,
      each: function (el) {
        results.push(el);
      },
      done: function (count) {
        resolve(results);
      },
    });
  });
}

function setSearch(s) {
  let el = document.getElementById("search");
  el.value = s;
}

document.addEventListener("DOMContentLoaded", (event) => {
  const eventList = document.getElementById("event-list");
  const search = document.getElementById("search");
  if (search && eventList) {
    eventList.addEventListener("htmx:afterSettle", async function (evt) {
      switch (evt.detail.target.id) {
        case "event-list":
          // Initial events loaded (on tab switch or search query change).
          await highlightResults(evt.detail.elt);
          window.scrollTo({ top: 0, behavior: "smooth" });
          break;

        case "load-more":
          // Infinite scroll loaded more events.
          // Highlight the added event.
          await highlightResults(evt.detail.elt);
          break;
      }
    });
  }
});
