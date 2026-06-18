const dateFormatter = new Intl.DateTimeFormat(navigator.language, {
  dateStyle: "long",
  timeStyle: "short",
});

function localizeDates(eventList) {
  eventList.querySelectorAll(".localize-date").forEach((el) => {
    el.textContent = dateFormatter.format(
      new Date(el.getAttribute("datetime")),
    );
  });
}

document.addEventListener("DOMContentLoaded", () => {
  const eventList = document.getElementById("event-list");
  if (eventList) {
    // Localize dates on all events on the first HTML load.
    localizeDates(eventList);

    eventList.addEventListener("htmx:afterSettle", function (evt) {
      // Localize dates on dynamically loaded events.
      localizeDates(evt.detail.elt);
    });
  }
});
