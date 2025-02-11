function changeTab(link) {
  // Clear search.
  const search = document.getElementById("search");
  if (search) {
    search.value = "";
  }

  // Make all links inactive.
  document.querySelectorAll(".nav-link").forEach(function (el) {
    el.classList.add("text-gray-400");
    el.classList.add("hover:text-amber-600");
    el.classList.remove("text-amber-600");

    el.parentElement.classList.remove("-mb-px");
    el.parentElement.classList.remove("border-l");
    el.parentElement.classList.remove("border-t");
    el.parentElement.classList.remove("border-r");
    el.parentElement.classList.remove("rounded-t");

    el.removeAttribute("aria-current");
  });

  // Make this link active.
  link.classList.remove("text-gray-400");
  link.classList.remove("hover:text-amber-600");
  link.classList.add("text-amber-600");

  link.parentElement.classList.add("-mb-px");
  link.parentElement.classList.add("border-l");
  link.parentElement.classList.add("border-t");
  link.parentElement.classList.add("border-r");
  link.parentElement.classList.add("rounded-t");

  link.setAttribute("aria-current", "page");
}
