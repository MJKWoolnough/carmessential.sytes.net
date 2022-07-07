import {amendNode, clearNode} from './lib/dom.js';
import {br, button, datalist, div, h1, input, li, option, span, textarea, ul} from './lib/html.js';
import {NodeArray, NodeMap, node, stringSort} from './lib/nodes.js';
import {registerPage, setPage} from './pages.js';
import {labels} from './shared.js';
import {ready, rpc} from './rpc.js';

class Treatment {
	#id: number;
	#name: string;
	#nameSpan: HTMLSpanElement;
	#group: string;
	#price: number;
	#description: string;
	#duration: number;
	[node]: HTMLLIElement;
	constructor(id = -1, name = "", group = "", price = 0, description = "", duration = 1) {
		this.#id = id;
		this[node] = li([
			this.#nameSpan = span(this.#name = name)
		]);
		this.#group = group;
		this.#price = price;
		this.#description = description;
		this.#duration = duration;
	}
	get id() {
		return this.#id;
	}
	set id(i: number) {
		this.#id = i;
	}
	get name() {
		return this.#name;
	}
	set name(n: string) {
		clearNode(this.#nameSpan, this.#name = n);
	}
	get group() {
		return this.#group;
	}
	get price() {
		return this.#price;
	}
	get description() {
		return this.#description;
	}
	get duration() {
		return this.#duration;
	}
}

type TreatmentNode = Treatment & {
	[node]: HTMLLIElement;
}

type Group = {
	group: string;
	arr: NodeArray<TreatmentNode, HTMLUListElement>;
	[node]: HTMLUListElement;
}

const treatmentSort = (a: Treatment, b: Treatment) => stringSort(a.name, b.name),
      contents = div();

ready.then(() => rpc.listTreatments().then(treatments => {
	const groups = new NodeMap<string, Group>(ul(), (a, b) => stringSort(a.group, b.group)),
	      groupList = datalist({"id": "groupNames"}),
	      treatmentTitle = h1(),
	      treatmentName = input({"type": "text"}),
	      treatmentGroup = input({"type": "text", "list": "groupNames"}),
	      treatmentPrice = input({"type": "number", "step": "0.01", "min": 0}),
	      treatmentDescription = textarea(),
	      treatmentDuration = input({"type": "number", "step": 1, "min": 1, "value": 1}),
	      submitTreatment = button({"onclick": function(this: HTMLButtonElement) {
		if (!treatmentName.value) {
			alert("Need a Name");
		} else if (!treatmentGroup.value) {
			alert("Need a Group");
		} else if (!treatmentPrice.value) {
			alert("Need a Price");
		} else if (!treatmentDescription.value) {
			alert("Need a Description");
		} else if (!treatmentDuration.value) {
			alert("Need a Duration");
		} else {
			const price = Math.floor(parseFloat(treatmentPrice.value) * 100),
			      duration = parseInt(treatmentDuration.value);
			amendNode(submitTreatment, {"disabled": true});
			(currTreatment.id === -1 ? rpc.addTreatment(treatmentName.value, treatmentGroup.value, price, treatmentDescription.value, duration).then(id => {
				currTreatment.id = id;
				addTreatment(currTreatment);
			}) : rpc.setTreatment(currTreatment.id, treatmentName.value, treatmentGroup.value, price, treatmentDescription.value, duration))
			.then(() => setPage("treatments"))
			.catch(err => alert("Error: " + err))
			.finally(() => amendNode(submitTreatment, {"disabled": false}));
		}
	      }}),
	      setTreatment = (treatment: Treatment) => {
		 treatmentName.value = treatment.name;
		 treatmentGroup.value = treatment.group;
		 treatmentPrice.value = (treatment.price / 100) + "";
		 treatmentDescription.value = treatment.description;
		 treatmentDuration.value = treatment.duration + "";
		 currTreatment = treatment;
		 clearNode(treatmentTitle, treatment.id === -1 ? "Create Treatment" : "Edit Treatment");
		 clearNode(submitTreatment, treatment.id === -1 ? "Create Treatment" : "Edit Treatment");
		 setPage("setTreatment");
	      },
	      addTreatment = (treatment: Treatment) => {
		if (!groups.has(treatment.group)) {
			const arr = new NodeArray<TreatmentNode, HTMLUListElement>(ul(), treatmentSort);
			amendNode(groupList, option({"value": treatment.group}));
			groups.set(treatment.group, {
				arr,
				"group": treatment.group,
				[node]: arr[node]
			});
		}
		groups.get(treatment.group)?.arr.push(treatment);
	      };
	let currTreatment: Treatment;
	for (const [id, name, group, price, description, duration]  of treatments) {
		addTreatment(new Treatment(id, name, group, price, description, duration));
	}
	amendNode(contents, [
		button({"onclick": () => setTreatment(new Treatment())}, "New Treatment"),
		groups[node]
	]);
	registerPage("setTreatment", "", div([
		treatmentTitle,
		labels("Treatment Name: ", treatmentName),
		br(),
		groupList,
		labels("Treatment Group: ", treatmentGroup),
		br(),
		labels("Treatment Price (£): ", treatmentPrice),
		br(),
		labels("Treatment Description: ", treatmentDescription),
		br(),
		labels("Treatment Duration (m): ", treatmentDuration),
		br(),
		submitTreatment
	]), () => {
		if (currTreatment.name !== treatmentName.value || currTreatment.group !== treatmentGroup.value || currTreatment.price !== parseFloat(treatmentPrice.value) * 100 || currTreatment.description !== treatmentDescription.value || currTreatment.duration !== parseInt(treatmentDuration.value)) {
			if (!confirm("There are unsaved changes, are you sure you wish to change page?")) {
				return Promise.reject();
			}
		}
		return Promise.resolve();
	});
}));

registerPage("treatments", "Edit Treatments", contents);
