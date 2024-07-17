document.addEventListener("DOMContentLoaded", function () {
	const fileDrop = document.getElementById("scan-form-file-drop");
	const fileInput = document.getElementById("scan-form-file-input");
	const nameInput = document.getElementById("scan-form-name-input");
	const rulesetInput = document.getElementById("scan-form-ruleset-input");
	const startButton = document.getElementById("scan-form-start-button");

	// Add drag and drop functionality to the file drop area
	fileDrop.addEventListener("dragover", (event) => {
		event.preventDefault();
		fileDrop.classList.add("dragging");
	});
	fileDrop.addEventListener("dragleave", () => {
		fileDrop.classList.remove("dragging");
	});
	fileDrop.addEventListener("drop", (event) => {
		event.preventDefault();
		fileDrop.classList.remove("dragging");
		const files = event.dataTransfer.files;
		if (files.length > 0) {
			fileInput.files = files;
			// Only use the first file
			fileDrop.querySelector("p").textContent = files[0].name;
			updateButtonState();
		}
	});

	// Add click functionality to the file drop area
	fileDrop.addEventListener("click", () => {
		fileInput.click();
	});

	// Update the file drop area when a file is selected
	fileInput.addEventListener("change", () => {
		if (fileInput.files.length > 0) {
			// Only use the first file
			fileDrop.querySelector("p").textContent = fileInput.files[0].name;
			updateButtonState();
		}
	});

	// Update the start button state when the form is updated
	function updateButtonState() {
		const isFormValid =
			fileInput.files.length > 0 &&
			nameInput.value.trim() !== "" &&
			rulesetInput.value !== "";
		startButton.disabled = !isFormValid;
	}

	// Listen for input changes
	nameInput.addEventListener("input", updateButtonState);
	rulesetInput.addEventListener("change", updateButtonState);

	// Prevent unfinished scans from being clicked
	document.querySelectorAll("a.scan-unfinished").forEach((link) => {
		link.addEventListener("click", function (event) {
			event.preventDefault();
		});
	});
});
