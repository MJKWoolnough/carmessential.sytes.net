import {amendNode, clearNode} from './lib/dom.js';
import {div, h1, li, ul} from './lib/html.js';
import {NodeMap, node, stringSort} from './lib/nodes.js';
import {ready} from './rpc.js';

type Page = {
	id: string;
	fn?: () => Promise<void>;
	[node]: HTMLLIElement;
}

export let header = "",
footer = "";

let currPage = "";

const body = div(),
      section = div(),
      pages = new NodeMap<string, Page>(ul(), (a: Page, b: Page) => stringSort(a.id, b.id));

export const setHeaderFooter = (h: string, f: string) => {
	header = h;
	footer = f;
	document.documentElement.innerHTML = `${h}<div id="ADMINBODY"></div>${f}`;
	document.getElementById("ADMINBODY")!.replaceWith(body);
	document.title = "Admin";
},
registerPage = (id: string, title: string, contents: HTMLElement, onchange?: () => Promise<void>) => {
	pages.set(id, {
		id,
		fn: onchange,
		[node]: li({"onclick": () => {
			if (currPage !== id) {
				(pages.get(currPage)?.fn?.() ?? Promise.resolve()).then(() => {
					clearNode(section, contents);
					currPage = id;
				});
			}
		}}, title)
	});
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
}).then(() => {
	amendNode(body, [
		pages[node],
		section
	]);
});
