import {amendNode, clearNode} from './lib/dom.js';
import {br, button, div, input, li, textarea, ul} from './lib/html.js';
import {NodeArray, NodeMap, node, stringSort} from './lib/nodes.js';
import {registerPage, setPage} from './pages.js';
import {labels} from './shared.js';
import {ready, rpc} from './rpc.js';

type Treatment = {
	id: number;
	name: string;
	price: number;
	description: string;
	duration: number;
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
	      treatmentName = input({"type": "text"}),
	      treatmentPrice = input({"type": "number", "step": "0.01", "min": 0}),
	      treatmentDescription = textarea(),
	      treatmentDuration = input({"type": "number", "step": 1, "min": 1, "value": 1}),
	      submitTreatment = button({"onclick": function(this: HTMLButtonElement) {
		if (!treatmentName.value) {
			alert("Need a Name");
		} else if (!treatmentPrice.value) {
			alert("Need a Price");
		} else if (!treatmentDescription.value) {
			alert("Need a Description");
		} else if (!treatmentDuration.value) {
			alert("Need a Duration");
		} else {
			const t: Treatment = {
				"id": currTreatment.id,
				"name": treatmentName.value,
				"price": Math.floor(parseFloat(treatmentPrice.value) * 100),
				"description": treatmentDescription.value,
				"duration": parseInt(treatmentDuration.value)
			      };
			(t.id === -1 ? rpc.addTreatment(t.name, "", t.price, t.description, t.duration).then(id => t.id = id) : rpc.setTreatment(t.id, t.name, "", t.price, t.description, t.duration)).then(() => currTreatment = t);
		}
	      }}, "Create Treatment"),
	      noTreatment = {
		"id": -1,
		"name": "",
		"price": 0,
		"description": "",
		"duration": 1
	      },
	      setTreatment = (treatment: Treatment = noTreatment) => {
		 treatmentName.value = treatment.name;
		 treatmentPrice.value = (treatment.price / 100) + "";
		 treatmentDescription.value = treatment.description;
		 treatmentDuration.value = treatment.duration + "";
		 currTreatment = treatment;
		 clearNode(submitTreatment, treatment.id === -1 ? "Create Treatment" : "Edit Treatment");
		 setPage("setTreatment");
	      };
	let currTreatment: Treatment = noTreatment;
	for (const [id, name, group, price, description, duration]  of treatments) {
		if (!groups.has(group)) {
			const arr = new NodeArray<TreatmentNode, HTMLUListElement>(ul(), treatmentSort);
			groups.set(group, {
				arr,
				group,
				[node]: arr[node]
			});
		}
		groups.get(group)?.arr.push({
			id,
			name, 
			price,
			description,
			duration,
			[node]: li(name)
		});
	}
	amendNode(contents, [
		button({"onclick": () => setTreatment()}, "New Treatment"),
		groups[node]
	]);
	registerPage("setTreatment", "", div([
		labels("Treatment Name: ", treatmentName),
		br(),
		labels("Treatment Price (£): ", treatmentPrice),
		br(),
		labels("Treatment Description: ", treatmentDescription),
		br(),
		labels("Treatment Duration (m): ", treatmentDuration),
		br(),
		submitTreatment
	]), () => {
		if (currTreatment.name !== treatmentName.value || currTreatment.price !== parseFloat(treatmentPrice.value) * 100 || currTreatment.description !== treatmentDescription.value || currTreatment.duration !== parseInt(treatmentDuration.value)) {
			if (!confirm("There are unsaved changes, are you sure you wish to change page?")) {
				return Promise.reject();
			}
		}
		return Promise.resolve();
	});
}));

registerPage("treatments", "Edit Treatments", contents);
