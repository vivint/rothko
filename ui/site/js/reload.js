var nonce = "";

window.setInterval(function() {
	var xhr = new XMLHttpRequest();
	xhr.open('GET', '/api/nonce');
	xhr.onload = function() {
		if (xhr.status !== 200) {
			return;
		}
		if (nonce !== "" && xhr.responseText !== nonce) {
			location.reload(true);
		}
		nonce = xhr.responseText;
	};
	xhr.send();
}, 1000);
