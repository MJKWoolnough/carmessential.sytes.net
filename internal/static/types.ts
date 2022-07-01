export type RPCType = {
	setHeaderFooter: (header: string, footer: string) => Promise<void>;
	listTreatments: () => Promise<[number, string, string, number, string, number][]>;
	addTreatment: (name: string, group: string, price: number, description: string, duration: number) => Promise<number>;
	setTreatment: (id: number, name: string, group: string, price: number, description: string, duration: number) => Promise<void>;
	removeTreatment: (id: number) => Promise<void>;
}
