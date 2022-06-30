import {WS} from './lib/conn.js';
import {clearNode} from './lib/dom.js';
import {div} from './lib/html.js';
import {RPC} from './lib/rpc.js';

declare const pageLoad: Promise<void>;

export let header = "",
footer = "";

const setHeaderFooter = (h: string, f: string) => {
	header = h;
	footer = f;
	document.documentElement.innerHTML = `${h}<div id="ADMINBODY"></div>${f}`;
	document.getElementById("ADMINBODY")!.replaceWith(body);
	document.title = "Admin";
};

export const rpc = {} as {
	setHeaderFooter: (header: string, footer: string) => Promise<void>;
},
body = div(),
ready = pageLoad.then(() => {
	clearNode(document.body, body);
	return WS("/admin")
}).then(ws => {
	const arpc = new RPC(ws);
	return arpc.await(-2).then(({header: h, footer: f}: {header: string, footer: string}) => {
		setHeaderFooter(h, f);
		return arpc.await(-1).then(() => {
			Object.freeze(Object.assign(rpc, {
				"setHeaderFooter": (header: string, footer: string) => arpc.request("setHeaderFooter", [header, footer]).finally(() => setHeaderFooter(header, footer))
			}));
		});
	});
});
