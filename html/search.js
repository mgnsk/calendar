/* global Mark */

const punct = ":;.,-–—‒_(){}[]!'\"+=".split("");

function highlightResults(targetNode, searchValue) {
  const mark = new Mark(targetNode);
  const s = searchValue.replace(/"+/g, ""); // Remove double quotes.

  mark.mark(s, {
    ignorePunctuation: punct,
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

  // Reset search value on page load.
  if (search) {
    search.value = "";
  }

  if (search && eventList) {
    highlightResults(eventList, search.value);

    eventList.addEventListener("htmx:afterSettle", function (evt) {
      highlightResults(evt.target, search.value);
    });
  }
});
