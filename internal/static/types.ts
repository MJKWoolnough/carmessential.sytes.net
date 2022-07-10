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

export type RPCType = {
	setHeaderFooter: (header: string, footer: string) => Promise<void>;
	listTreatments: () => Promise<[number, string, string, number, string, number][]>;
	addTreatment: (name: string, group: string, price: number, description: string, duration: number) => Promise<number>;
	setTreatment: (id: number, name: string, group: string, price: number, description: string, duration: number) => Promise<void>;
	removeTreatment: (id: number) => Promise<void>;
	getOrderTime: (id: number) => Promise<number>;
	addOrder: (bookings: Booking[]) => Promise<number[]>;
	removeOrder: (id: number) => Promise<void>;
	listBookings: (start: number, end: number) => Promise<Booking[]>;
	updateBooking: (booking: Booking) => Promise<void>;
	removeBooking: (id: number) => Promise<void>;
}
