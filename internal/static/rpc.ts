import {WS} from './lib/conn.js';
import {RPC} from './lib/rpc.js';

declare const pageLoad: Promise<void>;

export let header = "",
footer = "";

export const rpc = {} as {
},
ready = pageLoad.then(() => WS("/admin")).then(ws => {
	const arpc = new RPC(ws);
	return arpc.await(-2).then(({header: h, footer: f}: {header: string, footer: string}) => {
		header = h;
		footer = f;
		return arpc.await(-1).then(() => {
			Object.freeze(Object.assign(rpc, {
			}));
		});
	});
});
