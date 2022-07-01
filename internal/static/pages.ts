import {div} from './lib/html.js';

export let header = "",
footer = "";

export const setHeaderFooter = (h: string, f: string) => {
	header = h;
	footer = f;
	document.documentElement.innerHTML = `${h}<div id="ADMINBODY"></div>${f}`;
	document.getElementById("ADMINBODY")!.replaceWith(body);
	document.title = "Admin";
},
body = div();
