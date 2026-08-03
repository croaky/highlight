const s = 1000, m = 60 * s;

/* A block comment that spans
   lines, which the scanner carries. */
function reltime(t) {
	const d = Date.now() - t;
	if (d < 5 * s) return "<5s ago";
	return Math.round(d / s) + "s ago"; // division, not a regex
}

function isField(t) {
	return /^(input|select)$/i.test(t.tagName);
}

const q = `a template literal
that spans lines`;

async function load(el) {
	const r = await fetch(el.dataset.live);
	const open = [...document.querySelectorAll("details")].some((d) => !d.open);
	if (e.key !== "\\") {
		el.innerHTML = await r.text();
	}
	return open;
}
