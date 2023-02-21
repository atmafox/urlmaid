const suppList = document.getElementById("suppList");
const urlToTidy = document.getElementById("urlToTidy");
const tidyURLButton = document.getElementById("tidyURL");
const tidiedURLText = document.getElementById("tidiedURL");

const tidyURLFile = '/api/v1/tidy'

async function tidyURL() {
	var tidyURLData = {
		type: suppList.value,
		url: urlToTidy.value
	};

	const response = await fetch(tidyURLFile, {
		method: "POST",
		headers: {
			"Accept": "text/plain",
			"Content-Type": "application/json; charset=utf-8"
		},
		keepalive: false,
		body: JSON.stringify(tidyURLData)
	});

	if (!response.ok) {
		const message = `An error occurred: ${response.status}`;
		throw new Error(message);
	}

	tidiedURLText.innerHTML = await response.text();
}

tidyURLButton.addEventListener("click", () => tidyURL(), false);

