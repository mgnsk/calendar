let eventList = document.getElementById("event-list");
let search = document.getElementById("search");

if (search && eventList) {
  let mark = new Mark(eventList);
  eventList.addEventListener("htmx:afterSettle", function (evt) {
    let s = search.value.replace(/['"]+/g, "");
    mark.mark(s);
  });
}

function setSearch(s) {
  let el = document.getElementById("search");
  el.value = s;
  el.dispatchEvent(new Event("change"));
}
