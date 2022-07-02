import type {RPCType} from './types.js';
import {WS} from './lib/conn.js';
import {RPC} from './lib/rpc.js';
import {setHeaderFooter} from './pages.js';

declare const pageLoad: Promise<void>;

export const rpc = {} as RPCType,
ready = pageLoad.then(() => WS("/admin")).then(ws => {
	const arpc = new RPC(ws);
	return arpc.await(-2).then(({header: h, footer: f}: {header: string, footer: string}) => {
		setHeaderFooter(h, f);
		return arpc.await(-1).then(() => {
			Object.freeze(Object.assign(rpc, {
				"setHeaderFooter": (header: string, footer: string) => arpc.request("setHeaderFooter", [header, footer]).finally(() => setHeaderFooter(header, footer)),
				"listTreatments": () => arpc.request("listTreatments"),
				"addTreatment": (name: string, group: string, price: number, description: string, duration: number) => arpc.request("addTreatment", {name, group, price, description, duration}),
				"setTreatment": (id: number, name: string, group: string, price: number, description: string, duration: number) => arpc.request("addTreatment", {id, name, group, price, description, duration}),
				"removeTreatment": (id: number) => arpc.request("removeTreatment", id)
			}));
		});
	});
});
