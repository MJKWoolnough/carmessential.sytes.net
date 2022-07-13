import type {Booking, RPCType} from './types.js';
import {WS} from './lib/conn.js';
import {RPC} from './lib/rpc.js';
import {setHeaderFooter} from './pages.js';

declare const pageLoad: Promise<void>;

export const rpc = {} as RPCType,
ready = pageLoad.then(() => WS("/admin")).then(ws => {
	const arpc = new RPC(ws);
	return arpc.await(-2).then(([h, f]: [string, string]) => {
		setHeaderFooter(h, f);
		return arpc.await(-1).then(() => {
			Object.freeze(Object.assign(rpc, {
				"setHeaderFooter": (header: string, footer: string) => arpc.request("setHeaderFooter", [header, footer]).finally(() => setHeaderFooter(header, footer)),
				"listTreatments": () => arpc.request("listTreatments"),
				"addTreatment": (name: string, group: string, price: number, description: string, duration: number) => arpc.request("addTreatment", {name, group, price, description, duration}),
				"setTreatment": (id: number, name: string, group: string, price: number, description: string, duration: number) => arpc.request("setTreatment", {id, name, group, price, description, duration}),
				"removeTreatment": (id: number) => arpc.request("removeTreatment", id),
				"getOrderTime": (id: number) => arpc.request("getOrderTime", id),
				"addOrder": (bookings: Booking[]) => arpc.request("addOrder", bookings),
				"removeOrder": (id: number) => arpc.request("removeOrder", id),
				"listBookings": (start: number, end: number) => arpc.request("listBookings", [start, end]),
				"updateBooking": (b: Booking) => arpc.request("updateBooking", b),
				"removeBooking": (id: number) => arpc.request("removeBooking", id)
			}));
		});
	});
});
