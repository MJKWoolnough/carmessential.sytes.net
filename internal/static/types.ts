export type Booking = {
	id: number;
	date: number;
	blockNum: number;
	totalBlocks: number;
	treatmentID: number;
	name: string;
	emailAddress: string;
	phoneNumber: string;
	orderID: number;
}

export type Treatment = {
	id: number;
	name: string;
	group: string;
	price: number;
	description: string;
	duration: number;
}

export type Voucher = {
	id: number;
	code: string;
	name: string;
	expiry: number;
	isValue: boolean;
	value: number;
	valid: boolean;
}

export type OrderResponse = {
	orderID: number;
	bookings: number[];
	vouchers: [number, string][];
}

export type RPCType = {
	setHeaderFooter: (header: string, footer: string) => Promise<void>;
	listTreatments: () => Promise<Treatment[]>;
	addTreatment: (name: string, group: string, price: number, description: string, duration: number) => Promise<number>;
	setTreatment: (id: number, name: string, group: string, price: number, description: string, duration: number) => Promise<void>;
	removeTreatment: (id: number) => Promise<void>;
	getOrderTime: (id: number) => Promise<number>;
	addOrder: (bookings: Omit<Booking, "id" | "orderID">[], vouchers: Omit<Voucher, "id" | "code" | "valid" | "orderID">[]) => Promise<OrderResponse>;
	removeOrder: (id: number) => Promise<void>;
	listBookings: (start: number, end: number) => Promise<Booking[]>;
	updateBooking: (booking: Booking) => Promise<void>;
	removeBooking: (id: number) => Promise<void>;
	getVoucher: (id: number) => Promise<Voucher>;
	getVoucherByCode: (code: string) => Promise<Voucher>;
	updateVoucher: (id: number, name: string, expiry: number) => Promise<void>;
	removeVoucher: (id: number) => Promise<void>;
	setVoucherValid: (id: number, valid: boolean) => Promise<void>;
}
