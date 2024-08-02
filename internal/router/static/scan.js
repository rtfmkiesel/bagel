document.addEventListener("DOMContentLoaded", function () {
	// Add a listener to the search input
	const searchInput = document.getElementById("scan-results-search-input");
	searchInput.addEventListener("input", function () {
		filterResultsWithSearch();
	});

	// Add filter toggle button
	const filterToggleButton = document.getElementById(
		"scan-results-filter-toggle"
	);
	filterToggleButton.addEventListener("click", function () {
		toggleFilterList();
	});
	document.addEventListener("keydown", function (event) {
		if (event.key === "F" && event.shiftKey) {
			toggleFilterList();
		}
	});

	// Add filter reset button
	const filterResetButton = document.getElementById(
		"scan-results-filter-reset"
	);
	filterResetButton.addEventListener("click", function () {
		document.querySelectorAll(".scan-results-filter-element-selected").forEach((el) => {
			el.classList.remove("scan-results-filter-element-selected");
		});
		filterResultsWithFilters();
	});

	// Get all vuln classes and paths, populate the filter list
	populateFilters();

	// Add filter logic to filter elements
	addFilterListeners();
});

// Toggle the filter list
function toggleFilterList() {
	const filterDiv = document.getElementById("scan-results-filters");
	filterDiv.style.display =
		filterDiv.style.display === "block" ? "none" : "block";
}

// Filter results based on search input
function filterResultsWithSearch() {
	const searchString = document
		.getElementById("scan-results-search-input")
		.value.toLowerCase();

	document.querySelectorAll(".scan-result").forEach((result) => {
		const vulnClassText = result
			.querySelector("h3")
			.innerText.toLowerCase();
		const pathText = result
			.querySelector(".scan-result-data-path")
			.innerText.toLowerCase();
		const messageText = result
			.querySelector(".scan-result-data-message")
			.innerText.toLowerCase();

		const matchesSearch =
			vulnClassText.includes(searchString) ||
			pathText.includes(searchString) ||
			messageText.includes(searchString);

		if (matchesSearch) {
			result.style.display = "";
		} else {
			result.style.display = "none";
		}
	});
}

// Populate the filter list with elements
function populateFilters() {
	const vulnClassList = document.getElementById(
		"scan-results-filter-vulnclass"
	);
	const pathList = document.getElementById("scan-results-filter-path");

	let vulnClasses = new Set();
	let paths = new Set();

	const results = document
		.getElementById("scan-results")
		.getElementsByClassName("scan-result");
	for (let result of results) {
		// Get vuln classes
		const vulnClassElement = result.querySelector("h3");
		if (vulnClassElement) {
			// Can be multiple classes separated by commas
			vulnClassElement.textContent
				.split(",")
				.forEach((vc) => vulnClasses.add(vc.trim()));
		}

		// Get path
		const pathElement = result.querySelector(".scan-result-data-path");
		if (pathElement) {
			paths.add(pathElement.textContent.trim());
		}
	}

	// Sort all sets alphabetically
	vulnClasses = Array.from(vulnClasses).sort();
	paths = Array.from(paths).sort();

	createFilterElements(vulnClassList, vulnClasses, "vulnclass");
	createFilterElements(pathList, paths, "path");
}

// Creates a filter element (li containing an a) and appends it to the given element
function createFilterElements(appendTo, items, filterName) {
	items.forEach((item) => {
		const link = document.createElement("a");
		link.classList.add("scan-results-filter-element");
		link.classList.add(`filter-${filterName}`);
		link.innerText = item;

		const li = document.createElement("li");
		li.appendChild(link);
		appendTo.appendChild(li);
	});
}

// Binds filter logic (adds listener to filter elements)
function addFilterListeners() {
	const filterElements = document.querySelectorAll(
		".scan-results-filter-element"
	);
	filterElements.forEach((element) => {
		element.addEventListener("click", function () {
			element.classList.toggle("scan-results-filter-element-selected");
			filterResultsWithFilters();
		});
	});
}

// Filter results based on active filters
function filterResultsWithFilters() {
	const activeVulnClasses = Array.from(
		document.querySelectorAll(
			".filter-vulnclass.scan-results-filter-element-selected"
		)
	).map((el) => el.innerText);
	const activePaths = Array.from(
		document.querySelectorAll(
			".filter-path.scan-results-filter-element-selected"
		)
	).map((el) => el.innerText);

	document.querySelectorAll(".scan-result").forEach((result) => {
		const vulnClassText = result.querySelector("h3").innerText;
		const pathText = result.querySelector(
			".scan-result-data-path"
		).innerText;

		const matchesVulnClass =
			activeVulnClasses.length === 0 ||
			activeVulnClasses.some((vc) => vulnClassText.includes(vc));
		const matchesPath =
			activePaths.length === 0 || activePaths.includes(pathText);

		if (matchesVulnClass && matchesPath) {
			result.style.display = "";
		} else {
			result.style.display = "none";
		}
	});
}
