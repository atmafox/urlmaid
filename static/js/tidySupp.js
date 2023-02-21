const tidySuppList = document.getElementById("suppList");
tidySuppList.length = 0;
const defaultOption = new Option("Choose tidy method", 'defaultOption', true, true);
tidySuppList.add(defaultOption);

const suppURL = '/api/v1/supported'

async function tidySuppFill() {
	const response = await fetch(suppURL, {
		method: "GET",
		headers: {
			"Accept": "application/json"
		},
		keepalive: false
	});

	if (!response.ok) {
		const message = `An error occurred: ${response.status}`;
		throw new Error(message);
	}

	let option;
	const data = await response.json();

	for (let i = 0; i < data.length; i++) {
		option = new Option(data[i], data[i], false, false);
		tidySuppList.add(option);
	}
};

tidySuppFill().catch(error => {
	error.message; // 'An error occurred: 404'
});

