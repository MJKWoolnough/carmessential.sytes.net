import {amendNode, clearNode} from './lib/dom.js';
import {div, h1, li, style, ul} from './lib/html.js';
import {NodeMap, node, stringSort} from './lib/nodes.js';
import {ready} from './rpc.js';

type Page = {
	id: string;
	contents: HTMLElement;
	fn?: () => Promise<void>;
	[node]: HTMLLIElement | Comment;
}

export let header = "",
footer = "";

let currPage = "",
    css = `
#adminMenu {
	list-style: none;
	padding: 0;
}

#adminMenu > li {
	background-color: #fff;
	color: #000;
	cursor: pointer;
	display: inline-block;
	padding: 0 1em;
	text-align: center;
}

#adminMenu > li:hover {
	background-color: #000;
	color: #fff;
}
`;

const body = div(),
      section = div(),
      pages = new NodeMap<string, Page>(ul({"id": "adminMenu"}), (a: Page, b: Page) => stringSort(a.id, b.id));

export const setHeaderFooter = (h: string, f: string) => {
	header = h;
	footer = f;
	document.documentElement.innerHTML = `${h}<div id="ADMINBODY"></div>${f}`;
	document.getElementById("ADMINBODY")!.replaceWith(body);
	document.title = "Admin";
	amendNode(document.head, style({"type": "text/css"}, css));
},
registerPage = (id: string, title: string, contents: HTMLElement, onchange?: () => Promise<void>) => pages.set(id, {
	id,
	contents,
	fn: onchange,
	[node]: title ? li({"onclick": () => setPage(id)}, title) : document.createComment("")
}),
addCSS = (style: string) => {
	css += style;
},
setPage = (id: string) => {
	if (currPage !== id) {
		const page = pages.get(id);
		if (page) {
			(pages.get(currPage)?.fn?.() ?? Promise.resolve()).then(() => {
				clearNode(section, page.contents);
				currPage = id;
			});
		};
	}
};

ready.catch(e => {
	clearNode(body, [
		h1("Error"),
		div(e.message ?? "Unknown Error")
	]);
	if (!body.parentNode) {
		clearNode(document.body, body);
	}
	throw e;
}).then(() => amendNode(body, [
	pages[node],
	section
]));
