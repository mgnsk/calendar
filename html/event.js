function localizeDates(eventList) {
  eventList.querySelectorAll(".localize-date").forEach((el) => {
    const date = new Date(el.getAttribute("datetime"));
    if (!isNaN(date.getTime())) {
      el.textContent = date.toLocaleString(navigator.language, {
        dateStyle: "long",
        timeStyle: "short",
      });
    }
  });
}

document.addEventListener("DOMContentLoaded", () => {
  const eventList = document.getElementById("event-list");
  if (eventList) {
    localizeDates(eventList);

    eventList.addEventListener("htmx:afterSettle", function (evt) {
      localizeDates(evt.target);
    });
  }
});
